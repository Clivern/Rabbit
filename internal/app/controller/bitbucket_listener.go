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

// BitbucketListener controller
func BitbucketListener(c *gin.Context, messages chan<- string) {

	rawBody, _ := c.GetRawData()
	body := string(rawBody)

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	pushEvent := pkg.BitbucketPushEvent{}
	ok, err := pushEvent.LoadFromJSON(rawBody)

	if err != nil {
		logger.Info(fmt.Sprintf(
			`Invalid event received %s`,
			body,
		), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

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

	if len(pushEvent.Push.Changes) <= 0 || !pushEvent.Push.Changes[0].Created || pushEvent.Push.Changes[0].New.Type != "tag" {
		c.JSON(http.StatusOK, gin.H{
			"status": "Nice, Skip!",
		})
		return
	}

	// Push event received
	href := strings.ReplaceAll(
		viper.GetString("integrations.bitbucket.https_format"),
		"[.RepoFullName]",
		pushEvent.Repository.FullName,
	)

	if viper.GetString("integrations.bitbucket.clone_with") == "ssh" {
		href = strings.ReplaceAll(
			viper.GetString("integrations.bitbucket.ssh_format"),
			"[.RepoFullName]",
			pushEvent.Repository.FullName,
		)
	}

	releaseRequest := model.ReleaseRequest{
		Name:    pushEvent.Repository.Name,
		URL:     href,
		Version: pushEvent.Push.Changes[0].New.Name,
	}

	passToWorker(
		c,
		messages,
		logger,
		releaseRequest,
	)
}
