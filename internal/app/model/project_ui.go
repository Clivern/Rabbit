// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package model

// BinaryUI struct
type BinaryUI struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

// ReleaseUI struct
type ReleaseUI struct {
	Binaries []BinaryUI `json:"binaries"`
}

// ProjectUI struct
type ProjectUI struct {
	Name           string               `json:"name"`
	LatestRelease  string               `json:"latest_release"`
	CurrentRelease string               `json:"current_release"`
	Releases       map[string]ReleaseUI `json:"releases"`
}
