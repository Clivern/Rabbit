// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/spf13/viper"
)

// HealthCheck do a health check
func HealthCheck() {

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
		panic(report)
	} else {
		logger.Info("Health Check: I am Ok")
		fmt.Println("I am Ok")
	}
}
