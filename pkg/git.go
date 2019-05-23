// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"gopkg.in/src-d/go-git.v4"
)

// Repository struct
type Repository struct {
	URL string
}

// Clone Clones a Repository
func (r *Repository) Clone(directory string) (bool, error) {
	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL: r.URL,
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
