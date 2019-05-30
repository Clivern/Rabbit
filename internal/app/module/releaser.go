// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/pkg"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

// GoReleaserConfig config file for go releaser
const GoReleaserConfig = ".goreleaser.yml"

// GoReleaserChecksums checksums file
const GoReleaserChecksums = "checksums.txt"

// ReleaseRequest struct
type ReleaseRequest struct {
	Name    string
	URL     string
	Version string
}

// ReleasePath struct
type ReleasePath struct {
	VCS        string
	Author     string
	Repository string
}

// Releaser struct
type Releaser struct {
	ID             string
	RepositoryName string
	RepositoryURL  string
	Version        string
	LastCommit     string
	BuildPath      string
	LastTag        string
	ReleasePath    *ReleasePath
}

// NewReleaser returns a new instance of Releaser
func NewReleaser(repositoryName, repositoryURL, version string) (*Releaser, error) {
	correlation := hippo.NewCorrelation()
	releasePath, err := releasePathFromURL(repositoryURL)

	if err != nil {
		return nil, err
	}

	return &Releaser{
		ID:             correlation.UUIDv4(),
		RepositoryName: repositoryName,
		RepositoryURL:  repositoryURL,
		Version:        version,
		ReleasePath:    releasePath,
	}, nil
}

// releasePathFromURL
func releasePathFromURL(repositoryURL string) (*ReleasePath, error) {
	releasePath := &ReleasePath{}

	newRepositoryURL := repositoryURL

	// Parse github repo url
	if strings.Contains(newRepositoryURL, "github.com") {
		releasePath.VCS = "github.com"
		newRepositoryURL = strings.ReplaceAll(newRepositoryURL, "git@github.com:", "")
		newRepositoryURL = strings.ReplaceAll(newRepositoryURL, "https://github.com/", "")
		newRepositoryURL = strings.ReplaceAll(newRepositoryURL, ".git", "")

		if strings.Contains(newRepositoryURL, "/") {
			namespaces := strings.Split(newRepositoryURL, "/")
			releasePath.Author = strings.ToLower(namespaces[0])
			releasePath.Repository = strings.ToLower(namespaces[1])
		} else {
			return nil, fmt.Errorf("Unable to parse repository url [%s]", repositoryURL)
		}
	}

	return releasePath, nil
}

// Release build & release all binaries
func (r *Releaser) Release() (bool, error) {

	if r.BuildPath == "" || !hippo.DirExists(r.BuildPath) {
		return false, fmt.Errorf("Unable to find build path [%s]", r.BuildPath)
	}

	// Release with goreleaser if .goreleaser.yml file exists
	if hippo.FileExists(fmt.Sprintf("%s/%s", r.BuildPath, GoReleaserConfig)) {
		return r.releaseWithGoReleaser()
	}

	// Manually release it or create goreleaser file & release

	return true, nil
}

// releaseWithGoReleaser release a project
func (r *Releaser) releaseWithGoReleaser() (bool, error) {
	cmd := pkg.NewShellCommand()

	_, err := cmd.Exec(r.BuildPath, "goreleaser", "--snapshot", "--skip-publish", "--rm-dist")

	if err != nil {
		return false, err
	}

	var files []string

	distPath := fmt.Sprintf("%s/%s", r.BuildPath, "dist")

	err = filepath.Walk(distPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".gz" && filepath.Ext(path) != ".txt" {
			return nil
		}
		files = append(files, filepath.Base(path))
		return nil
	})

	if err != nil {
		return false, err
	}

	// Validate checksums
	for _, file := range files {
		if filepath.Ext(file) != ".gz" {
			continue
		}

		res, err := pkg.ValidateChecksum(
			fmt.Sprintf("%s/%s", distPath, file),
			fmt.Sprintf("%s/%s", distPath, GoReleaserChecksums),
		)
		if err != nil {
			return false, err
		}
		if !res {
			return false, fmt.Errorf("File [%s] checksum is not valid", fmt.Sprintf("%s/%s", distPath, file))
		}
	}

	newPath := fmt.Sprintf(
		"%s/%s/%s/%s/%s",
		strings.TrimSuffix(viper.GetString("releases.path"), "/"),
		r.ReleasePath.VCS,
		r.ReleasePath.Author,
		r.ReleasePath.Repository,
		r.Version,
	)

	err = os.MkdirAll(newPath, 0777)

	if err != nil {
		return false, err
	}

	for _, file := range files {
		err = os.Rename(
			fmt.Sprintf("%s/%s", distPath, file),
			fmt.Sprintf("%s/%s", newPath, strings.ToLower(file)),
		)
		if err != nil {
			return false, err
		}
	}

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

// LoadFromJSON load object from json
func (c *ReleaseRequest) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &c)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON converts object to json
func (c *ReleaseRequest) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
