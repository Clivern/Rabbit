// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/model"
	"github.com/clivern/rabbit/pkg"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"strings"
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

	if ok {

		// Ping event received
		if eventName == "ping" {
			var pingEvent pkg.PingEvent
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

		// Create event received
		if eventName == "create" {
			var createEvent pkg.CreateEvent
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
				"{$full_name}",
				createEvent.Repository.FullName,
			)

			if viper.GetString("integrations.github.clone_with") == "ssh" {
				href = strings.ReplaceAll(
					viper.GetString("integrations.github.ssh_format"),
					"{$full_name}",
					createEvent.Repository.FullName,
				)
			}

			releaseRequest := model.ReleaseRequest{
				Name:    createEvent.Repository.Name,
				URL:     href,
				Version: createEvent.Ref,
			}

			validate := pkg.Validator{}

			if validate.IsEmpty(releaseRequest.Name) {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"error":  "Repository name is required",
				})
				return
			}

			if validate.IsEmpty(releaseRequest.URL) {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"error":  "Repository url is required",
				})
				return
			}

			if validate.IsEmpty(releaseRequest.Version) {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"error":  "Repository version is required",
				})
				return
			}

			requestBody, err := releaseRequest.ConvertToJSON()

			if err != nil {
				logger.Error(fmt.Sprintf(
					`Error while converting request body to json %s`,
					err.Error(),
				), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"error":  "Invalid request",
				})
				return
			}

			if viper.GetString("broker.driver") == "redis" {
				driver := hippo.NewRedisDriver(
					viper.GetString("redis.addr"),
					viper.GetString("redis.password"),
					viper.GetInt("redis.db"),
				)

				// connect to redis server
				ok, err = driver.Connect()

				if err != nil {
					logger.Error(fmt.Sprintf(
						`Unable to connect to redis server %s`,
						err.Error(),
					), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

					c.JSON(http.StatusServiceUnavailable, gin.H{
						"status": "error",
						"error":  "Internal server error",
					})
					return
				}

				if !ok {
					logger.Error(
						`Unable to connect to redis server`,
						zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")),
					)

					c.JSON(http.StatusServiceUnavailable, gin.H{
						"status": "error",
						"error":  "Internal server error",
					})
					return
				}

				// ping check
				ok, err = driver.Ping()

				if err != nil {
					logger.Error(fmt.Sprintf(
						`Unable to ping redis server %s`,
						err.Error(),
					), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

					c.JSON(http.StatusServiceUnavailable, gin.H{
						"status": "error",
						"error":  "Internal server error",
					})
					return
				}

				if !ok {
					logger.Error(
						`Unable to ping redis server`,
						zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")),
					)

					c.JSON(http.StatusServiceUnavailable, gin.H{
						"status": "error",
						"error":  "Internal server error",
					})
					return
				}

				logger.Info(fmt.Sprintf(
					`Send request [%s] to workers`,
					requestBody,
				), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

				driver.Publish(
					viper.GetString("broker.redis.channel"),
					requestBody,
				)
			} else {
				logger.Info(fmt.Sprintf(
					`Send request [%s] to workers`,
					requestBody,
				), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

				messages <- requestBody
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "Nice!",
		})
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
	}
}
