// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// ValidateChecksum validates a checksum
func ValidateChecksum(file, checksumsFile string) (bool, error) {
	b, err := ioutil.ReadFile(checksumsFile)
	if err != nil {
		return false, err
	}

	checkSums := string(b)

	checksum, err := GetChecksum(file)

	if err != nil {
		return false, err
	}

	return strings.Contains(checkSums, checksum), nil
}

// GetChecksum get a checksum
func GetChecksum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
