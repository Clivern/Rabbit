// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/clivern/rabbit/internal/app/module"
	"github.com/clivern/rabbit/pkg"
)

// ReleasePackage clones and build binaries
func ReleasePackage(repositoryName, repositoryURL, repositoryTag string) {

	if repositoryName == "" || repositoryURL == "" || repositoryTag == "" {
		panic("Please provide the repository name, URL and tag")
	}

	s := pkg.NewSpinner("%s Hold on...")
	s.Start()
	defer s.Stop()

	releaser, err := module.NewReleaser(repositoryName, repositoryURL, repositoryTag)

	if err != nil {
		panic(err.Error())
	}

	_, err = releaser.Clone()

	if err != nil {
		panic(err.Error())
	}

	_, err = releaser.Release()

	if err != nil {
		panic(err.Error())
	}

	_, err = releaser.Cleanup()

	if err != nil {
		panic(err.Error())
	}
}
