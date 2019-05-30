// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/clivern/rabbit/internal/app/module"
)

// ChanWorker runs async jobs
func ChanWorker(messages <-chan string) {
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
