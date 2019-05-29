// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/clivern/rabbit/internal/app/module"
)

// ReleasePackage clones and build binaries
func ReleasePackage(repositoryName, repositoryURL, repositoryTag string) {

	if repositoryName == "" || repositoryURL == "" || repositoryTag == "" {
		panic("Please provide the repository name, URL and tag")
	}

	releaser := module.NewReleaser(repositoryName, repositoryURL, repositoryTag)
	fmt.Println(releaser.Clone())
	fmt.Println(releaser)
	fmt.Println(releaser.Cleanup())
}
