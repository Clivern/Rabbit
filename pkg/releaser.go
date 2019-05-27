// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

// GoReleaserURL the go package URL
const GoReleaserURL = "github.com/goreleaser/goreleaser"

// Releaser struct
type Releaser struct {
}

// NewReleaser creates a new releaser
func NewReleaser() *Releaser {
	return &Releaser{}
}

// Install installs goreleaser
func (r *Releaser) Install(path string) {
	shell := NewShellCommand()
	shell.Exec(path, "go", "get", "-u", GoReleaserURL)
}

// Release builds binaries
func (r *Releaser) Release(path string) {
	shell := NewShellCommand()
	shell.Exec(path, "goreleaser", "--snapshot", "--skip-publish", "--rm-dist")
}
