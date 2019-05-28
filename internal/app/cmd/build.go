// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/clivern/rabbit/internal/app/module"
)

// ReleasePackage clones and build binaries
func ReleasePackage() {
	releaser := module.NewReleaser("hippo", "https://github.com/Clivern/Hippo.git", "v1.3.0")
	fmt.Println(releaser.Clone())
	fmt.Println(releaser)
	fmt.Println(releaser.Cleanup())
}
