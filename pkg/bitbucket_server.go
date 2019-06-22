// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// BitbucketServerWebhookParser struct
type BitbucketServerWebhookParser struct {
	UserAgent    string
	HubSignature string
	Headers      map[string]string
	Body         string
}

// LoadFromJSON update object from json
func (e *BitbucketServerWebhookParser) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &e)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON convert object to json
func (e *BitbucketServerWebhookParser) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetUserAgent sets user agent
func (e *BitbucketServerWebhookParser) SetUserAgent(userAgent string) {
	e.UserAgent = userAgent
}

// GetUserAgent gets user agent
func (e *BitbucketServerWebhookParser) GetUserAgent() string {
	return e.UserAgent
}

// SetHubSignature sets hub signature
func (e *BitbucketServerWebhookParser) SetHubSignature(hubSignature string) {
	e.HubSignature = hubSignature
}

// GetHubSignature gets hub signature
func (e *BitbucketServerWebhookParser) GetHubSignature() string {
	return e.HubSignature
}

// SetBody sets body
func (e *BitbucketServerWebhookParser) SetBody(body string) {
	e.Body = body
}

// GetBody gets body
func (e *BitbucketServerWebhookParser) GetBody() string {
	return e.Body
}

// SetHeader sets header
func (e *BitbucketServerWebhookParser) SetHeader(key string, value string) {
	e.Headers[key] = value
}

// GetHeader gets header
func (e *BitbucketServerWebhookParser) GetHeader(key string) string {
	return e.Headers[key]
}

// VerifySignature verify signature
func (e *BitbucketServerWebhookParser) VerifySignature(secret string) bool {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(e.Body))

	return fmt.Sprintf("sha256=%s", hex.EncodeToString(hash.Sum(nil))) == e.HubSignature
}
