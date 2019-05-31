// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/module"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Worker runs async jobs
func Worker(workerID int, messages <-chan string) {

	workerName := fmt.Sprintf("Worker#%d", workerID)

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	logger.Info(fmt.Sprintf(
		`%s started`,
		workerName,
	), zap.String("WorkerName", workerName))

	if viper.GetString("broker.driver") == "redis" {

		driver := hippo.NewRedisDriver(
			viper.GetString("broker.redis.addr"),
			viper.GetString("broker.redis.password"),
			viper.GetInt("broker.redis.db"),
		)

		// connect to redis server
		ok, err := driver.Connect()

		if err != nil {
			logger.Error(fmt.Sprintf(
				"Error while connecting to redis server [%s] [%s]",
				viper.GetString("broker.redis.addr"),
				err.Error(),
			))
			panic(err.Error())
		}

		if !ok {
			logger.Error(fmt.Sprintf(
				"Unable to connect to redis server [%s] [%s]",
				viper.GetString("broker.redis.addr"),
				workerName,
			))
			panic(fmt.Sprintf(
				"Unable to connect to redis server [%s] [%s]",
				viper.GetString("broker.redis.addr"),
				workerName,
			))
		}

		// ping check
		ok, err = driver.Ping()

		if err != nil {
			logger.Error(fmt.Sprintf(
				"Error while pinging redis server [%s] [%s]",
				viper.GetString("broker.redis.addr"),
				err.Error(),
			))
			panic(err.Error())
		}

		if !ok {
			logger.Error(fmt.Sprintf(
				"Unable to ping redis server [%s] [%s]",
				viper.GetString("broker.redis.addr"),
				workerName,
			))
			panic(fmt.Sprintf(
				"Unable to ping redis server [%s] [%s]",
				viper.GetString("broker.redis.addr"),
				workerName,
			))
		}

		driver.Subscribe(viper.GetString("broker.redis.channel"), func(message hippo.Message) error {
			var releaseRequest module.ReleaseRequest

			ok, err := releaseRequest.LoadFromJSON([]byte(message.Payload))

			if !ok || err != nil {
				logger.Error(fmt.Sprintf(
					"Error while parsing message payload [%s]",
					message.Payload,
				))
				return nil
			}

			logger.Info(fmt.Sprintf(
				"Init releaser for package [%s] url [%s] version [%s]",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			releaser, err := module.NewReleaser(
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			)

			defer releaser.Cleanup()

			if err != nil {
				logger.Error(fmt.Sprintf(
					"Error while parsing data for package [%s] url [%s] version [%s]: [%s]",
					releaseRequest.Name,
					releaseRequest.URL,
					releaseRequest.Version,
					err.Error(),
				))
				return nil
			}

			logger.Info(fmt.Sprintf(
				"Cloning package [%s] url [%s] version [%s]",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			_, err = releaser.Clone()

			if err != nil {
				logger.Error(fmt.Sprintf(
					"Error while cloning package [%s] url [%s] version [%s]: [%s]",
					releaseRequest.Name,
					releaseRequest.URL,
					releaseRequest.Version,
					err.Error(),
				))
				return nil
			}

			logger.Info(fmt.Sprintf(
				"Releasing package [%s] url [%s] version [%s]",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			_, err = releaser.Release()

			if err != nil {
				logger.Error(fmt.Sprintf(
					"Error while releasing package [%s] url [%s] version [%s]: [%s]",
					releaseRequest.Name,
					releaseRequest.URL,
					releaseRequest.Version,
					err.Error(),
				))
				return nil
			}

			logger.Info(fmt.Sprintf(
				"Package [%s] url [%s] version [%s] released, do cleanup",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			return nil
		})
	} else {
		for message := range messages {
			var releaseRequest module.ReleaseRequest

			ok, err := releaseRequest.LoadFromJSON([]byte(message))

			if !ok || err != nil {
				logger.Error(fmt.Sprintf(
					"Error while parsing message payload [%s]",
					message,
				))
				continue
			}

			logger.Info(fmt.Sprintf(
				"Init releaser for package [%s] url [%s] version [%s]",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			releaser, err := module.NewReleaser(
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			)

			if err != nil {
				logger.Error(fmt.Sprintf(
					"Error while parsing data for package [%s] url [%s] version [%s]: [%s]",
					releaseRequest.Name,
					releaseRequest.URL,
					releaseRequest.Version,
					err.Error(),
				))
				continue
			}

			logger.Info(fmt.Sprintf(
				"Cloning package [%s] url [%s] version [%s]",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			_, err = releaser.Clone()

			if err != nil {
				logger.Error(fmt.Sprintf(
					"Error while cloning package [%s] url [%s] version [%s]: [%s]",
					releaseRequest.Name,
					releaseRequest.URL,
					releaseRequest.Version,
					err.Error(),
				))
				releaser.Cleanup()
				continue
			}

			logger.Info(fmt.Sprintf(
				"Releasing package [%s] url [%s] version [%s]",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			_, err = releaser.Release()

			if err != nil {
				logger.Error(fmt.Sprintf(
					"Error while releasing package [%s] url [%s] version [%s]: [%s]",
					releaseRequest.Name,
					releaseRequest.URL,
					releaseRequest.Version,
					err.Error(),
				))
				releaser.Cleanup()
				continue
			}

			logger.Info(fmt.Sprintf(
				"Package [%s] url [%s] version [%s] released, do cleanup",
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			))

			releaser.Cleanup()
		}
	}
}
