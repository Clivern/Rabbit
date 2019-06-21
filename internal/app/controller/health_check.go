// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"net/http"

	"github.com/clivern/hippo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// HealthCheck controller
func HealthCheck(c *gin.Context) {

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	healthChecker := hippo.NewHealthChecker()

	healthChecker.AddCheck("redis_check", func() (bool, error) {
		return hippo.RedisCheck(
			"redis_service",
			viper.GetString("redis.addr"),
			viper.GetString("redis.password"),
			viper.GetInt("redis.db"),
		)
	})

	healthChecker.RunChecks()

	if healthChecker.ChecksStatus() == "DOWN" {
		report, _ := healthChecker.ChecksReport()
		logger.Error("Health Check Error: " + report)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "Internal server error",
		})
	} else {
		logger.Info("Health Check: I am Ok")
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
