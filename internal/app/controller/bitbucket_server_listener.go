// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/model"
	"github.com/clivern/rabbit/pkg"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// BitbucketServerListener controller
func BitbucketServerListener(c *gin.Context, messages chan<- string) {
	rawBody, _ := c.GetRawData()
	body := string(rawBody)

	parser := &pkg.BitbucketServerWebhookParser{
		UserAgent:    c.GetHeader("User-Agent"),
		HubSignature: c.GetHeader("X-Hub-Signature"),
		Body:         body,
	}

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	ok := parser.VerifySignature(viper.GetString("integrations.bitbucket_server.webhook_secret"))

	if !ok {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	pushEvent := pkg.BitbucketServerPushEvent{}
	ok, err := pushEvent.LoadFromJSON(rawBody)

	if err != nil {
		logger.Info(
			fmt.Sprintf(
				`Invalid event received %s, error: %s`,
				body,
				err,
			),
			zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")),
		)

		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	if !ok {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	if len(pushEvent.Changes) <= 0 || pushEvent.GetTag() == "" {
		c.JSON(http.StatusOK, gin.H{
			"status": "Nice, Skip!",
		})
		return
	}

	repoFullName := fmt.Sprintf(
		"%s/%s",
		strings.ToLower(pushEvent.Repository.Project.Key),
		strings.ToLower(pushEvent.Repository.Slug),
	)

	href := strings.ReplaceAll(
		viper.GetString("integrations.bitbucket_server.https_format"),
		"[.RepoFullName]",
		repoFullName,
	)

	if viper.GetString("integrations.bitbucket_server.clone_with") == "ssh" {
		href = strings.ReplaceAll(
			viper.GetString("integrations.bitbucket_server.ssh_format"),
			"[.RepoFullName]",
			repoFullName,
		)
	}

	logger.Info(fmt.Sprintf("Clone URL is: %s", href))

	releaseRequest := model.ReleaseRequest{
		Name:    pushEvent.Repository.Name,
		URL:     href,
		Version: pushEvent.GetTag(),
	}

	passToWorker(
		c,
		messages,
		logger,
		releaseRequest,
	)
}
