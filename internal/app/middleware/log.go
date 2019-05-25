// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"fmt"
	"github.com/clivern/hippo"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)

// Logger middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request
		var bodyBytes []byte
		t := time.Now()

		// Workaround for issue https://github.com/gin-gonic/gin/issues/1651
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		logger, _ := hippo.NewLogger("debug", "json", []string{"stdout"})
		logger.Info(
			fmt.Sprintf(
				`Incoming request %s %s %s`,
				c.Request.Method,
				c.Request.URL,
				string(bodyBytes),
			),
			zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")),
		)

		c.Next()

		// after request
		latency := time.Since(t)
		status := c.Writer.Status()
		size := c.Writer.Size()

		logger.Info(
			fmt.Sprintf(
				`Outgoing response code %d, size %d time spent %s`,
				status,
				size,
				latency,
			),
			zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")),
		)
	}
}
