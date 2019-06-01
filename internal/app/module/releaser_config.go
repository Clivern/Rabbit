// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"fmt"
	"os"
	"strings"
)

// GoReleaserConfig config file for go releaser
const GoReleaserConfig = ".goreleaser.yml"

// GoReleaserChecksums checksums file
const GoReleaserChecksums = "checksums.txt"

// GoReleaserConfigTemplate config file template
const GoReleaserConfigTemplate = `# This is an example goreleaser.yaml file with some sane defaults.
# Make sure to check the documentation at http://goreleaser.com
project_name: {project_name}
before:
  hooks:
    # you may remove this if you don't use vgo
    - go mod download
    # you may remove this if you don't need go generate
    - go generate ./...
builds:
- env:
  - CGO_ENABLED=0
archives:
- replacements:
    darwin: darwin
    linux: linux
    windows: windows
    386: i386
    amd64: x86_64
checksum:
  name_template: 'checksums.txt'
snapshot:
  name_template: "{snapshot.name_template}"
changelog:
  sort: asc
  filters:
    exclude:
    - '^docs:'
    - '^test:'
`

// CreateReleaserConfig creates a .goreleaser.yml config file
func CreateReleaserConfig(path, projectName, releaseName string) (bool, error) {
	content := strings.ReplaceAll(
		GoReleaserConfigTemplate,
		"{snapshot.name_template}",
		releaseName,
	)
	content = strings.ReplaceAll(
		content,
		"{project_name}",
		projectName,
	)

	f, err := os.Create(fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(path, "/"),
		GoReleaserConfig,
	))
	if err != nil {
		return false, err
	}

	defer f.Close()

	_, err = f.WriteString(content)

	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteReleaserConfig deletes .goreleaser.yml config file
func DeleteReleaserConfig(path string) error {
	return os.Remove(fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(path, "/"),
		GoReleaserConfig,
	))
}
