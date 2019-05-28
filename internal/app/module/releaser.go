// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/pkg"
	"github.com/spf13/viper"
	"strings"
)

// Releaser struct
type Releaser struct {
	ID             string
	RepositoryName string
	RepositoryURL  string
	Version        string
	LastCommit     string
	BuildPath      string
	LastTag        string
}

// NewReleaser returns a new instance of Releaser
func NewReleaser(repositoryName, repositoryURL, version string) *Releaser {
	correlation := hippo.NewCorrelation()

	return &Releaser{
		ID:             correlation.UUIDv4(),
		RepositoryName: repositoryName,
		RepositoryURL:  repositoryURL,
		Version:        version,
	}
}

// Release build & release all binaries
func (r *Releaser) Release() (bool, error) {
	return true, nil
}

// Clone clones a repository
func (r *Releaser) Clone() (bool, error) {
	cmd := pkg.NewShellCommand()

	r.BuildPath = fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(viper.GetString("build.path"), "/"),
		r.ID,
	)

	_, err := cmd.Exec(strings.TrimSuffix(viper.GetString("build.path"), "/"), "git", "clone", r.RepositoryURL, r.ID)

	if err != nil {
		return false, err
	}

	result, err := cmd.Exec(r.BuildPath, "git", "rev-parse", "HEAD")

	if err != nil {
		return false, err
	}

	if result.Stdout == "" {
		return false, fmt.Errorf("Cloned repository %s is corrupted", r.RepositoryURL)
	}

	r.LastCommit = strings.TrimSpace(result.Stdout)

	_, err = cmd.Exec(r.BuildPath, "git", "fetch", "--tags")

	if err != nil {
		return false, err
	}

	_, err = cmd.Exec(r.BuildPath, "git", "checkout", fmt.Sprintf("tags/%s", r.Version))

	if err != nil {
		return false, err
	}

	result, err = cmd.Exec(r.BuildPath, "git", "describe", "--tags")

	if err != nil {
		return false, err
	}

	r.LastTag = strings.TrimSpace(result.Stdout)

	if r.LastTag != r.Version {
		return false, fmt.Errorf("Repository last tag [%s] and version [%s] not matching", r.LastTag, r.Version)
	}

	return true, nil
}

// Cleanup remove the build dir
func (r *Releaser) Cleanup() (bool, error) {
	if !hippo.DirExists(r.BuildPath) {
		return true, nil
	}

	cmd := pkg.NewShellCommand()
	_, err := cmd.Exec(strings.TrimSuffix(viper.GetString("build.path"), "/"), "rm", "-rf", r.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}
