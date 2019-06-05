// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/model"
	"github.com/clivern/rabbit/internal/app/module"
	"github.com/clivern/rabbit/pkg"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
)

// CreateRelease controller
func CreateRelease(c *gin.Context, messages chan<- string) {

	var releaseRequest model.ReleaseRequest
	validate := pkg.Validator{}

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	rawBody, err := c.GetRawData()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request",
		})
		return
	}

	ok, err := releaseRequest.LoadFromJSON(rawBody)

	if !ok || err != nil {
		logger.Info(fmt.Sprintf(
			`Invalid request body %s`,
			string(rawBody),
		), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request",
		})
		return
	}

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

	c.Status(http.StatusAccepted)
}

// GetReleaseByID controller
func GetReleaseByID(c *gin.Context) {

	var project *model.Project

	ID := c.Param("id")
	dataStore := &module.RedisDataStore{}
	status, err := dataStore.Connect()

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "Internal server error",
		})
		return
	}

	if !status {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "Internal server error",
		})
		return
	}

	status, err = dataStore.ProjectExistsByUUID(ID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	if !status {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project is not exist",
		})
		return
	}

	project, err = dataStore.GetProjectByUUID(ID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, project)
}
