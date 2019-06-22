// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/cmd"
	"github.com/clivern/rabbit/internal/app/controller"
	"github.com/clivern/rabbit/internal/app/middleware"
	"github.com/drone/envsubst"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {

	var exec string
	var configFile string

	flag.StringVar(&configFile, "config", "config.prod.yml", "config")
	flag.StringVar(&exec, "exec", "", "exec")
	flag.Parse()

	configUnparsed, err := ioutil.ReadFile(configFile)

	if err != nil {
		panic(fmt.Sprintf(
			"Error while reading config file [%s]: %s",
			configFile,
			err.Error(),
		))
	}

	configParsed, err := envsubst.EvalEnv(string(configUnparsed))

	if err != nil {
		panic(fmt.Sprintf(
			"Error while parsing config file [%s]: %s",
			configFile,
			err.Error(),
		))
	}

	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer([]byte(configParsed)))

	if err != nil {
		panic(fmt.Sprintf(
			"Error while loading configs [%s]: %s",
			configFile,
			err.Error(),
		))
	}

	if exec != "" {
		switch exec {
		case "health":
			cmd.HealthCheck()
		}
		return
	}

	if viper.GetString("log.output") != "stdout" {
		dir, _ := filepath.Split(viper.GetString("log.output"))
		if !hippo.DirExists(dir) {
			if _, err := hippo.EnsureDir(dir, 777); err != nil {
				panic(fmt.Sprintf(
					"Directory [%s] creation failed with error: %s",
					dir,
					err.Error(),
				))
			}
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

	if !hippo.DirExists(strings.TrimSuffix(viper.GetString("build.path"), "/")) {
		if _, err := hippo.EnsureDir(strings.TrimSuffix(viper.GetString("build.path"), "/"), 777); err != nil {
			panic(fmt.Sprintf(
				"Build directory [%s] creation failed with error: %s",
				strings.TrimSuffix(viper.GetString("build.path"), "/"),
				err.Error(),
			))
		}
	}

	if !hippo.DirExists(strings.TrimSuffix(viper.GetString("releases.path"), "/")) {
		if _, err := hippo.EnsureDir(strings.TrimSuffix(viper.GetString("releases.path"), "/"), 777); err != nil {
			panic(fmt.Sprintf(
				"Releases directory [%s] creation failed with error: %s",
				strings.TrimSuffix(viper.GetString("releases.path"), "/"),
				err.Error(),
			))
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

	messages := make(chan string, viper.GetInt("broker.native.capacity"))
	r := gin.Default()

	r.Use(middleware.Correlation())
	r.Use(middleware.Logger())

	r.StaticFS(
		"/releases",
		http.Dir(strings.TrimSuffix(viper.GetString("releases.path"), "/")),
	)
	r.GET("/", controller.Index)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.String(http.StatusNoContent, "")
	})
	r.GET("/_health", controller.HealthCheck)

	r.POST("/api/project", func(c *gin.Context) {
		controller.CreateProject(c, messages)
	})

	r.GET("/api/project/:id", controller.GetProjectByID)
	r.GET("/api/project", controller.GetProjects)

	r.POST(strings.TrimSuffix(viper.GetString("integrations.github.webhook_uri"), "/"), func(c *gin.Context) {
		controller.GithubListener(c, messages)
	})

	r.POST(strings.TrimSuffix(viper.GetString("integrations.bitbucket.webhook_uri"), "/"), func(c *gin.Context) {
		controller.BitbucketListener(c, messages)
	})

	r.POST(strings.TrimSuffix(viper.GetString("integrations.bitbucket_server.webhook_uri"), "/"), func(c *gin.Context) {
		controller.BitbucketServerListener(c, messages)
	})

	for i := 0; i < viper.GetInt("broker.native.workers"); i++ {
		go controller.Worker(i+1, messages)
	}

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
