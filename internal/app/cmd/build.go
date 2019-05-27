// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/clivern/rabbit/pkg"
	"github.com/spf13/viper"
)

// ReleasePackage clones and build binaries
func ReleasePackage() {
	releasesPath := viper.GetString("releases.path")

	cmd := pkg.NewShellCommand()

	fmt.Println("Run git clone https://github.com/Clivern/Beaver.git beaver")
	cmd.Exec(releasesPath, "git", "clone", "https://github.com/Clivern/Beaver.git", "beaver")

	fmt.Println("Start Releasing")
	re := pkg.NewReleaser()
	re.Install(releasesPath + "/beaver")
	re.Release(releasesPath + "/beaver")
}
