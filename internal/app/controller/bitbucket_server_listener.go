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

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

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

	if len(pushEvent.Changes) <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": "Nice, Skip!",
		})
		return
	}

	var cloneFormat string
	var href string

	// Push event received
	if viper.GetString("integrations.bitbucket_server.clone_with") == "ssh" {
		cloneFormat = viper.GetString("integrations.bitbucket_server.ssh_format")
	} else {
		cloneFormat = viper.GetString("integrations.bitbucket_server.https_format")
	}

	href = strings.ReplaceAll(
		cloneFormat,
		"{$project.key}",
		strings.ToLower(pushEvent.Repository.Project.Key),
	)

	href = strings.ReplaceAll(
		href,
		"{$repository.slug}",
		strings.ToLower(pushEvent.Repository.Slug),
	)

	logger.Info(fmt.Sprintf("Clone URL is: %s", href))

	for _, changes := range pushEvent.Changes {
		ref := changes.Ref
		if ref.Type == pkg.BitbucketServerChangeTypeTag {
			releaseRequest := model.ReleaseRequest{
				Name:    pushEvent.Repository.Name,
				URL:     href,
				Version: ref.DisplayID,
			}

			passToWorker(
				c,
				messages,
				logger,
				releaseRequest,
			)
		}
	}
}
