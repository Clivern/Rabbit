// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"bytes"
	"os/exec"
)

// Shell struct
type Shell struct {
}

// Result struct
type Result struct {
	Stdout string
	Stderr string
}

// NewShellCommand creates a new shell command
func NewShellCommand() *Shell {
	return &Shell{}
}

// Exec execute a command
func (s *Shell) Exec(path, name string, arg ...string) (*Result, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, arg...)

	cmd.Dir = path
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return &Result{}, err
	}

	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())

	return &Result{Stdout: outStr, Stderr: errStr}, nil
}
