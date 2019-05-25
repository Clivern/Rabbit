// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"github.com/clivern/hippo"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
)

// Correlation middleware
func Correlation() gin.HandlerFunc {
	return func(c *gin.Context) {
		corralationID := c.Request.Header.Get("X-Correlation-ID")

		logger, _ := hippo.NewLogger("debug", "json", []string{"stdout"})
		logger.Info(`Hello World!`, zap.String("CorrelationID", "123"))

		if strings.TrimSpace(corralationID) == "" {
			correlation := hippo.NewCorrelation()
			c.Request.Header.Add("X-Correlation-ID", correlation.UUIDv4())
		}
		c.Next()
	}
}
