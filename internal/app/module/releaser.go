// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/model"
	"github.com/clivern/rabbit/pkg"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Releaser struct
type Releaser struct {
	ID          string
	ProjectName string
	ProjectURL  string
	Version     string
	LastCommit  string
	BuildPath   string
	LastTag     string
	Binaries    map[string]string
}

// NewReleaser returns a new instance of Releaser
func NewReleaser(projectName, projectURL, version string) (*Releaser, error) {
	correlation := hippo.NewCorrelation()

	return &Releaser{
		ID:          correlation.UUIDv4(),
		ProjectName: projectName,
		ProjectURL:  projectURL,
		Version:     version,
	}, nil
}

// Release build & release all binaries
func (r *Releaser) Release() (bool, error) {

	if r.BuildPath == "" || !hippo.DirExists(r.BuildPath) {
		return false, fmt.Errorf("Unable to find build path [%s]", r.BuildPath)
	}

	releaseName := viper.GetString("releases.name")
	releaseName = strings.ReplaceAll(
		releaseName,
		"[.Tag]",
		strings.ToLower(r.Version),
	)

	ok, err := CreateReleaserConfig(r.BuildPath, strings.ToLower(r.ProjectName), releaseName)

	if err != nil || !ok {
		return false, fmt.Errorf(
			"Unable to create/update goreleaser config file [%s]",
			fmt.Sprintf("%s/%s", r.BuildPath, GoReleaserConfig),
		)
	}

	// double check the goreleaser file
	if !hippo.FileExists(fmt.Sprintf("%s/%s", r.BuildPath, GoReleaserConfig)) {
		return false, fmt.Errorf(
			"Unable to create/update goreleaser config file [%s]",
			fmt.Sprintf("%s/%s", r.BuildPath, GoReleaserConfig),
		)
	}

	result, err := r.Build()

	if err != nil {
		return false, err
	}

	if !result {
		return false, fmt.Errorf("Error while building project")
	}

	return r.Store()
}

// Build builds the project
func (r *Releaser) Build() (bool, error) {
	cmd := pkg.NewShellCommand()

	_, err := cmd.Exec(
		r.BuildPath,
		"goreleaser",
		"--snapshot",
		"--skip-publish",
		"--rm-dist",
		"--parallelism",
		strconv.Itoa(viper.GetInt("build.parallelism")),
	)

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

	if len(r.Binaries) <= 0 {
		r.Binaries = make(map[string]string)
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

		checksum, err := pkg.GetChecksum(fmt.Sprintf("%s/%s", distPath, file))

		if err != nil {
			return false, fmt.Errorf("Unable to get checksum for [%s]", fmt.Sprintf("%s/%s", distPath, file))
		}
		r.Binaries[checksum] = file
	}

	return true, nil
}

// Store stores the project
func (r *Releaser) Store() (bool, error) {
	var project *model.Project

	dataStore := &RedisDataStore{}
	status, err := dataStore.Connect()

	if err != nil {
		return false, err
	}

	if !status {
		return false, fmt.Errorf("Unable to connect to redis")
	}

	status, err = dataStore.ProjectExistsByURL(r.ProjectURL)

	if err != nil {
		return false, err
	}

	if status {
		project, err = dataStore.GetProjectByURL(r.ProjectURL)

		if err != nil {
			return false, err
		}
	} else {
		project = model.NewProject()
		project.SetName(r.ProjectName)
		project.SetUUID(r.ID)
		project.SetURL(r.ProjectURL)
	}

	newPath := fmt.Sprintf(
		"%s/%s/%s",
		strings.TrimSuffix(viper.GetString("releases.path"), "/"),
		project.GetUUID(),
		r.Version,
	)

	distPath := fmt.Sprintf("%s/%s", r.BuildPath, "dist")

	err = os.MkdirAll(newPath, 0777)

	if err != nil {
		return false, err
	}

	for checksum, fileName := range r.Binaries {
		err = os.Rename(
			fmt.Sprintf("%s/%s", distPath, fileName),
			fmt.Sprintf("%s/%s", newPath, strings.ToLower(fileName)),
		)
		if err != nil {
			return false, err
		}
		project.AddBinary(r.Version, fileName, checksum, "SHA256")
	}

	ok, err := dataStore.StoreProject(project)

	return ok, err
}

// Clone clones a repository
func (r *Releaser) Clone() (bool, error) {
	cmd := pkg.NewShellCommand()

	r.BuildPath = fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(viper.GetString("build.path"), "/"),
		r.ID,
	)

	_, err := cmd.Exec(strings.TrimSuffix(viper.GetString("build.path"), "/"), "git", "clone", r.ProjectURL, r.ID)

	if err != nil {
		return false, err
	}

	result, err := cmd.Exec(r.BuildPath, "git", "rev-parse", "HEAD")

	if err != nil {
		return false, err
	}

	if result.Stdout == "" {
		return false, fmt.Errorf("Cloned repository %s is corrupted", r.ProjectURL)
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

	_, err := cmd.Exec(
		strings.TrimSuffix(viper.GetString("build.path"), "/"),
		"rm",
		"-rf",
		r.ID,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}
