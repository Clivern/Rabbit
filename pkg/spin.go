// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"fmt"
	"sync/atomic"
	"time"
)

// ClearLine go to the beginning of the line and clear it
const ClearLine = "\r\033[K"

// Spinner types.
var (
	Spin1  = `|/-\`
	Spin2  = `◴◷◶◵`
	Spin3  = `◰◳◲◱`
	Spin4  = `◐◓◑◒`
	Spin5  = `▉▊▋▌▍▎▏▎▍▌▋▊▉`
	Spin6  = `▌▄▐▀`
	Spin7  = `╫╪`
	Spin8  = `■□▪▫`
	Spin9  = `←↑→↓`
	Spin10 = `⦾⦿`
	Spin11 = `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
	Spin12 = `⠋⠙⠚⠞⠖⠦⠴⠲⠳⠓`
	Spin13 = `⠄⠆⠇⠋⠙⠸⠰⠠⠰⠸⠙⠋⠇⠆`
	Spin14 = `⠋⠙⠚⠒⠂⠂⠒⠲⠴⠦⠖⠒⠐⠐⠒⠓⠋`
	Spin15 = `⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠴⠲⠒⠂⠂⠒⠚⠙⠉⠁`
	Spin16 = `⠈⠉⠋⠓⠒⠐⠐⠒⠖⠦⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈`
	Spin17 = `⠁⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈⠈`
)

// Spinner main type
type Spinner struct {
	frames []rune
	pos    int
	active uint64
	text   string
}

// NewSpinner Spinner with args
func NewSpinner(text string) *Spinner {
	s := &Spinner{
		text: ClearLine + text,
	}
	s.Set(Spin4)
	return s
}

// Set frames to the given string which must not use spaces.
func (s *Spinner) Set(frames string) {
	s.frames = []rune(frames)
}

// Start shows the spinner.
func (s *Spinner) Start() *Spinner {
	if atomic.LoadUint64(&s.active) > 0 {
		return s
	}
	atomic.StoreUint64(&s.active, 1)
	go func() {
		for atomic.LoadUint64(&s.active) > 0 {
			fmt.Printf(s.text, s.next())
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return s
}

// Stop hides the spinner.
func (s *Spinner) Stop() bool {
	if x := atomic.SwapUint64(&s.active, 0); x > 0 {
		fmt.Printf(ClearLine)
		return true
	}
	return false
}

func (s *Spinner) next() string {
	r := s.frames[s.pos%len(s.frames)]
	s.pos++
	return string(r)
}
