// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/controller"
	"github.com/clivern/rabbit/internal/app/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func main() {

	var configFile string

	flag.StringVar(&configFile, "config", "config.prod.yml", "config")
	flag.Parse()

	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Sprintf(
			"Error while loading config file [%s]: %s",
			configFile,
			err.Error(),
		))
	}

	if viper.GetString("log.output") != "stdout" {
		dir, _ := filepath.Split(viper.GetString("log.output"))
		if !hippo.DirExists(dir) {
			panic(fmt.Sprintf(
				"Logs output directory [%s] not exist",
				dir,
			))
		}

		if !hippo.FileExists(viper.GetString("log.output")) {
			f, err := os.Create(viper.GetString("log.output"))
			if err != nil {
				panic(fmt.Sprintf(
					"Error while creating log file [%s]: %s",
					viper.GetString("log.output"),
					err.Error(),
				))
			}
			defer f.Close()
		}
	}

	if viper.GetString("app.mode") == "prod" {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		if viper.GetString("log.output") == "stdout" {
			gin.DefaultWriter = os.Stdout
		} else {
			f, _ := os.Create(fmt.Sprintf("%s/gin.log", viper.GetString("log.output")))
			gin.DefaultWriter = io.MultiWriter(f)
		}
	}

	r := gin.Default()
	r.Use(middleware.Correlation())
	r.Use(middleware.Logger())
	r.GET("/", controller.Index)
	r.GET("/_health", controller.HealthCheck)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.String(http.StatusNoContent, "")
	})

	if viper.GetBool("app.tls.status") {
		r.RunTLS(
			fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
			viper.GetString("app.tls.pemPath"),
			viper.GetString("app.tls.keyPath"),
		)
	} else {
		r.Run(
			fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
		)
	}
}
