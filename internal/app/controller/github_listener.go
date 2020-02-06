// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/clivern/rabbit/internal/app/model"
	"github.com/clivern/rabbit/pkg"

	"github.com/clivern/hippo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// GithubListener controller
func GithubListener(c *gin.Context, messages chan<- string) {
	rawBody, _ := c.GetRawData()
	body := string(rawBody)

	parser := &pkg.GithubWebhookParser{
		UserAgent:      c.GetHeader("User-Agent"),
		GithubDelivery: c.GetHeader("X-GitHub-Delivery"),
		GitHubEvent:    c.GetHeader("X-GitHub-Event"),
		HubSignature:   c.GetHeader("X-Hub-Signature"),
		Body:           body,
	}

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	ok := parser.VerifySignature(viper.GetString("integrations.github.webhook_secret"))
	eventName := parser.GetGitHubEvent()

	if !ok && viper.GetString("integrations.github.webhook_secret") != "" {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	// Ping event received
	if eventName == "ping" {
		var pingEvent pkg.GithubPingEvent
		pingEvent.LoadFromJSON(rawBody)

		logger.Info(fmt.Sprintf(
			`Github event %s received`,
			eventName,
		), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

		c.JSON(http.StatusOK, gin.H{
			"status": "Nice!",
		})
		return
	}

	if eventName != "create" {
		c.JSON(http.StatusOK, gin.H{
			"status": "Nice, Skip!",
		})
		return
	}

	var createEvent pkg.GithubCreateEvent
	createEvent.LoadFromJSON(rawBody)

	logger.Info(fmt.Sprintf(
		`Github event %s received`,
		eventName,
	), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

	if createEvent.RefType != "tag" {
		logger.Info(fmt.Sprintf(
			`Github create event received but RefType is %s not tag`,
			createEvent.RefType,
		), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

		c.JSON(http.StatusOK, gin.H{
			"status": "Nice!",
		})
		return
	}

	href := strings.ReplaceAll(
		viper.GetString("integrations.github.https_format"),
		"[.RepoFullName]",
		createEvent.Repository.FullName,
	)

	if viper.GetString("integrations.github.clone_with") == "ssh" {
		href = strings.ReplaceAll(
			viper.GetString("integrations.github.ssh_format"),
			"[.RepoFullName]",
			createEvent.Repository.FullName,
		)
	}

	releaseRequest := model.ReleaseRequest{
		Name:    createEvent.Repository.Name,
		URL:     href,
		Version: createEvent.Ref,
	}

	passToWorker(
		c,
		messages,
		logger,
		releaseRequest,
	)
}
