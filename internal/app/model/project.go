// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package model

import (
	"encoding/json"
	"time"
)

// Binary struct
type Binary struct {
	FileName     string `json:"file_name"`
	Checksum     string `json:"checksum"`
	ChecksumType string `json:"checksum_type"`
}

// Release struct
type Release struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Binaries  []*Binary `json:"binaries"`
}

// Meta struct
type Meta struct {
}

// Project struct
type Project struct {
	Name     string              `json:"name"`
	UUID     string              `json:"uuid"`
	URL      string              `json:"url"`
	Releases map[string]*Release `json:"releases"`
	Meta     *Meta               `json:"meta"`
}

// NewProject creates a new project model
func NewProject() *Project {
	return &Project{}
}

// SetName sets a project name
func (p *Project) SetName(name string) {
	p.Name = name
}

// SetUUID sets a project UUID
func (p *Project) SetUUID(uuid string) {
	p.UUID = uuid
}

// SetURL sets a project URL
func (p *Project) SetURL(url string) {
	p.URL = url
}

// SetRelease sets a new release or update old one
func (p *Project) SetRelease(version string, createdAt time.Time) {
	if len(p.Releases) <= 0 {
		p.Releases = make(map[string]*Release)
	}
	p.Releases[version] = &Release{
		Version:   version,
		CreatedAt: createdAt,
	}
}

// DeleteRelease deletes a release
func (p *Project) DeleteRelease(version string) {
	if _, ok := p.Releases[version]; ok {
		delete(p.Releases, version)
	}
}

// AddBinary adds a binary to release
func (p *Project) AddBinary(version, fileName, checksum, checksumType string) {
	if _, ok := p.Releases[version]; !ok {
		p.SetRelease(version, time.Now())
	}
	p.Releases[version].Binaries = append(p.Releases[version].Binaries, &Binary{
		FileName:     fileName,
		Checksum:     checksum,
		ChecksumType: checksumType,
	})
}

// LoadFromJSON update object from json
func (p *Project) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &p)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON convert object to json
func (p *Project) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
