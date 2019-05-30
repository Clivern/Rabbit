// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

// Release controller
func Release(c *gin.Context) {
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
		panic(fmt.Errorf("Unable to connect to redis server [%s]", viper.GetString("redis.addr")))
	}

	// ping check
	ok, err = driver.Ping()

	if err != nil {
		panic(err.Error())
	}

	if !ok {
		panic(fmt.Errorf("Unable to connect to redis server [%s]", viper.GetString("redis.addr")))
	}

	driver.Publish(
		viper.GetString("redis.channel"),
		`{"repo":"https://github.com/Clivern/Hippo.git"}`,
	)

	c.Status(http.StatusAccepted)
}
