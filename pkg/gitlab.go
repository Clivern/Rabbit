// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"encoding/json"
)

// GitlabWebhookParser struct
type GitlabWebhookParser struct {
	GitlabEvent string
	GitlabToken string
	Headers     map[string]string
	Body        string
}

// LoadFromJSON update object from json
func (e *GitlabWebhookParser) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &e)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON convert object to json
func (e *GitlabWebhookParser) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetGitlabToken sets gitlab token
func (e *GitlabWebhookParser) SetGitlabToken(gitlabToken string) {
	e.GitlabToken = gitlabToken
}

// GetGitlabToken gets gitlab token
func (e *GitlabWebhookParser) GetGitlabToken() string {
	return e.GitlabToken
}

// SetGitlabEvent sets gitlab event
func (e *GitlabWebhookParser) SetGitlabEvent(gitlabEvent string) {
	e.GitlabEvent = gitlabEvent
}

// GetGitlabEvent gets gitlab event
func (e *GitlabWebhookParser) GetGitlabEvent() string {
	return e.GitlabEvent
}

// SetBody sets body
func (e *GitlabWebhookParser) SetBody(body string) {
	e.Body = body
}

// GetBody gets body
func (e *GitlabWebhookParser) GetBody() string {
	return e.Body
}

// SetHeader sets header
func (e *GitlabWebhookParser) SetHeader(key string, value string) {
	e.Headers[key] = value
}

// GetHeader gets header
func (e *GitlabWebhookParser) GetHeader(key string) string {
	return e.Headers[key]
}

// VerifySecret verifies the secret
func (e *GitlabWebhookParser) VerifySecret(secretToken string) bool {
	return e.GetGitlabToken() == secretToken
}
