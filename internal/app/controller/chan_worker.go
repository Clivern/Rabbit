// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
)

// ChanWorker runs async jobs
func ChanWorker(messages <-chan string) {
	for message := range messages {
		fmt.Printf("Received: %s\n", message)
	}
	fmt.Println("Done receiving!")
}
