// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/module"
	"github.com/spf13/viper"
)

// Worker runs async jobs
func Worker(messages <-chan string) {

	if viper.GetString("broker.driver") == "redis" {

		driver := hippo.NewRedisDriver(
			viper.GetString("redis.addr"),
			viper.GetString("redis.password"),
			viper.GetInt("redis.db"),
		)

		// connect to redis server
		ok, err := driver.Connect()

		if err != nil {
			panic(err.Error())
		}

		if !ok {
			panic(fmt.Errorf(
				"Unable to connect to redis server [%s]",
				viper.GetString("redis.addr"),
			))
		}

		// ping check
		ok, err = driver.Ping()

		if err != nil {
			panic(err.Error())
		}

		if !ok {
			panic(fmt.Errorf(
				"Unable to connect to redis server [%s]",
				viper.GetString("redis.addr"),
			))
		}

		driver.Subscribe("rabbit", func(message hippo.Message) error {
			var releaseRequest module.ReleaseRequest

			ok, err := releaseRequest.LoadFromJSON([]byte(message.Payload))

			if !ok || err != nil {
				return nil
			}

			releaser, err := module.NewReleaser(
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			)

			defer releaser.Cleanup()

			if err != nil {
				return nil
			}

			_, err = releaser.Clone()

			if err != nil {
				return nil
			}

			_, err = releaser.Release()

			if err != nil {
				return nil
			}
			return nil
		})
	} else {
		for message := range messages {
			var releaseRequest module.ReleaseRequest

			ok, err := releaseRequest.LoadFromJSON([]byte(message))

			if !ok || err != nil {
				continue
			}

			releaser, err := module.NewReleaser(
				releaseRequest.Name,
				releaseRequest.URL,
				releaseRequest.Version,
			)

			if err != nil {
				continue
			}

			_, err = releaser.Clone()

			if err != nil {
				releaser.Cleanup()
				continue
			}

			_, err = releaser.Release()

			if err != nil {
				releaser.Cleanup()
				continue
			}

			releaser.Cleanup()
		}
	}
}
