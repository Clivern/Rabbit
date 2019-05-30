// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

// Release struct
type Release struct {
}

// Repository struct
type Repository struct {
	ID             string
	RepositoryName string
	RepositoryURL  string
	Version        string
	Releases       []Release
}
