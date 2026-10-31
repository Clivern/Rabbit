// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package lb

import (
	"cmp"
	"errors"
	"hash/fnv"
	"math/rand/v2"
	"slices"
	"sync"
)

var (
	ErrEmpty    = errors.New("lb: no ids to pick from")
	ErrStrategy = errors.New("lb: unknown strategy")
)

const (
	Random                Strategy = "random"
	RoundRobin            Strategy = "round_robin"
	LeastOutstandingTasks Strategy = "least_outstanding_tasks"
	Sticky                Strategy = "sticky"
)

// Strategy is how Pick chooses among ids.
type Strategy string

// Options control how Pick chooses an id.
type Options struct {
	// Strategy defaults to Random.
	Strategy Strategy
	// Key counts round robin turns separately, for example per agent.
	Key string
	// StickyKey always picks the same id for the same key while the ids stay the same.
	StickyKey string
	// Outstanding is the open work per id for LeastOutstandingTasks; a missing id has none.
	Outstanding map[string]int
}

// Picker chooses one of a set of ids, such as the live instances of an agent.
type Picker struct {
	mu   sync.Mutex
	next map[string]int
	intn func(int) int
}

// New returns a picker.
func New() *Picker {
	return &Picker{
		next: make(map[string]int),
		intn: rand.IntN,
	}
}

// Pick returns the index in ids of the chosen id, so callers can index their own slice.
func (p *Picker) Pick(ids []string, opts Options) (int, error) {
	if len(ids) == 0 {
		return 0, ErrEmpty
	}

	if opts.StickyKey != "" {
		return slices.Index(ids, sticky(ids, opts.StickyKey)), nil
	}

	switch cmp.Or(opts.Strategy, Random) {
	case Random:
		return p.intn(len(ids)), nil
	case RoundRobin:
		return slices.Index(ids, p.roundRobin(opts.Key, ids)), nil
	case LeastOutstandingTasks:
		return p.leastOutstanding(ids, opts.Outstanding), nil
	case Sticky:
		// The sticky key already won above, so a missing key falls back to turns.
		return slices.Index(ids, p.roundRobin(opts.Key, ids)), nil
	}

	return 0, ErrStrategy
}

// roundRobin takes turns per key in sorted id order, so the order does not depend on how ids arrive.
func (p *Picker) roundRobin(key string, ids []string) string {
	ordered := slices.Sorted(slices.Values(ids))

	p.mu.Lock()
	i := p.next[key]
	p.next[key] = i + 1
	p.mu.Unlock()

	return ordered[i%len(ordered)]
}

// leastOutstanding returns the index of an id with the least open work, breaking ties at random.
func (p *Picker) leastOutstanding(ids []string, outstanding map[string]int) int {
	least := outstanding[ids[0]]
	for _, id := range ids[1:] {
		least = min(least, outstanding[id])
	}

	var tied []int
	for i, id := range ids {
		if outstanding[id] == least {
			tied = append(tied, i)
		}
	}

	return tied[p.intn(len(tied))]
}

// sticky hashes key onto the sorted ids.
func sticky(ids []string, key string) string {
	ordered := slices.Sorted(slices.Values(ids))
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))

	return ordered[int(h.Sum32())%len(ordered)]
}
