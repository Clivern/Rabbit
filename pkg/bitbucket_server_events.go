// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	bitbucketDateFormat string = "2006-01-02T15:04:05+0000"
	// BitbucketServerChangeTypeTag event change type
	BitbucketServerChangeTypeTag = "TAG"
)

// BitbucketServerPushEvent event
type BitbucketServerPushEvent struct {
	EventKey   string              `json:"eventKey"`
	Date       bitbucketServerDate `json:"date"`
	Actor      bitbucketUser       `json:"actor"`
	Repository struct {
		Slug         string `json:"slug"`
		ID           int    `json:"id"`
		Name         string `json:"name"`
		SCMID        string `json:"scmId"`
		State        string `json:"state"`
		StateMessage string `json:"statusMessage"`
		Forkable     bool   `json:"forkable"`
		Project      struct {
			Key   string        `json:"key"`
			ID    int           `json:"id"`
			Name  string        `json:"name"`
			Type  string        `json:"type"`
			Owner bitbucketUser `json:"owner"`
		} `json:"project"`
		Public bool `json:"public"`
	} `json:"repository"`
	Changes []struct {
		Ref struct {
			ID        string `json:"id"`
			DisplayID string `json:"displayId"`
			Type      string `json:"type"`
		} `json:"ref"`
		RefID    string `json:"refId"`
		FromHash string `json:"fromHash"`
		ToHash   string `json:"toHash"`
		Type     string `json:"type"`
	} `json:"changes"`
}

type bitbucketUser struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	ID           int    `json:"id"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	Slug         string `json:"slug"`
	Type         string `json:"type"`
}

type bitbucketServerDate struct {
	time.Time
}

// LoadFromJSON update object from json
func (bt *bitbucketServerDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		bt.Time = time.Now()
		return
	}
	bt.Time, err = time.Parse(bitbucketDateFormat, s)
	return
}

// LoadFromJSON update object from json
func (e *BitbucketServerPushEvent) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &e)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON convert object to json
func (e *BitbucketServerPushEvent) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
