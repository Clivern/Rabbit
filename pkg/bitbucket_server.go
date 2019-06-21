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

// BitbucketServer struct
type BitbucketServer struct {
	UserAgent    string
	HubSignature string
	Headers      map[string]string
	Body         string
}

// LoadFromJSON update object from json
func (e *BitbucketServer) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &e)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON convert object to json
func (e *BitbucketServer) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetUserAgent sets user agent
func (e *BitbucketServer) SetUserAgent(userAgent string) {
	e.UserAgent = userAgent
}

// GetUserAgent gets user agent
func (e *BitbucketServer) GetUserAgent() string {
	return e.UserAgent
}

// SetHubSignature sets hub signature
func (e *BitbucketServer) SetHubSignature(hubSignature string) {
	e.HubSignature = hubSignature
}

// GetHubSignature gets hub signature
func (e *BitbucketServer) GetHubSignature() string {
	return e.HubSignature
}

// SetBody sets body
func (e *BitbucketServer) SetBody(body string) {
	e.Body = body
}

// GetBody gets body
func (e *BitbucketServer) GetBody() string {
	return e.Body
}

// SetHeader sets header
func (e *BitbucketServer) SetHeader(key string, value string) {
	e.Headers[key] = value
}

// GetHeader gets header
func (e *BitbucketServer) GetHeader(key string) string {
	return e.Headers[key]
}

// VerifySignature verify signature
func (e *BitbucketServer) VerifySignature(secret string) bool {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(e.Body))

	return fmt.Sprintf("sha256=%s", hex.EncodeToString(hash.Sum(nil))) == e.HubSignature
}
