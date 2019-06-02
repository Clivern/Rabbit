// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package model

import (
	"encoding/json"
	"time"
)

// Release struct
type Release struct {
	UUID      string    `json:"uuid"`
	URL       string    `json:"url"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

// LoadFromJSON update object from json
func (e *Release) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &e)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON convert object to json
func (e *Release) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
