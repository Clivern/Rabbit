// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ChanRelease controller
func ChanRelease(c *gin.Context, messages chan<- string) {
	messages <- "Hello"

	c.Status(http.StatusAccepted)
}
