// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitPicker(t *testing.T) {
	a := "invoice-extractor-1"
	b := "invoice-extractor-2"
	c := "invoice-extractor-3"

	t.Run("empty ids", func(t *testing.T) {
		_, err := New().Pick(nil, Options{})
		assert.ErrorIs(t, err, ErrEmpty)
	})

	t.Run("single id", func(t *testing.T) {
		got, err := New().Pick([]string{a}, Options{})
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})

	t.Run("unknown strategy", func(t *testing.T) {
		_, err := New().Pick([]string{a}, Options{Strategy: "weighted"})
		assert.ErrorIs(t, err, ErrStrategy)
	})

	t.Run("random uses injected index", func(t *testing.T) {
		p := New()
		p.intn = func(n int) int {
			assert.Equal(t, 3, n)

			return 1
		}
		got, err := p.Pick([]string{a, b, c}, Options{Strategy: Random})
		require.NoError(t, err)
		assert.Equal(t, 1, got)
	})

	t.Run("round robin cycles in id order and returns indexes into the given ids", func(t *testing.T) {
		p := New()
		ids := []string{c, a, b}
		var picked []string
		for range 4 {
			got, err := p.Pick(ids, Options{Strategy: RoundRobin, Key: "invoice-extractor"})
			require.NoError(t, err)
			picked = append(picked, ids[got])
		}

		assert.Equal(t, []string{a, b, c, a}, picked)
	})

	t.Run("round robin counts turns per key", func(t *testing.T) {
		p := New()
		ids := []string{a, b}

		first, _ := p.Pick(ids, Options{Strategy: RoundRobin, Key: "one"})
		other, _ := p.Pick(ids, Options{Strategy: RoundRobin, Key: "two"})
		assert.Equal(t, first, other)
	})

	t.Run("least outstanding picks the lightest", func(t *testing.T) {
		p := New()
		p.intn = func(int) int { return 0 }
		got, err := p.Pick([]string{a, b}, Options{Strategy: LeastOutstandingTasks, Outstanding: map[string]int{a: 2}})
		require.NoError(t, err)
		assert.Equal(t, 1, got)
	})

	t.Run("least outstanding breaks ties at random", func(t *testing.T) {
		p := New()
		p.intn = func(n int) int {
			assert.Equal(t, 2, n)

			return 1
		}
		got, err := p.Pick([]string{a, b, c}, Options{Strategy: LeastOutstandingTasks, Outstanding: map[string]int{a: 2}})
		require.NoError(t, err)
		assert.Equal(t, 2, got)
	})

	t.Run("sticky key is stable whatever the order", func(t *testing.T) {
		p := New()
		first, err := p.Pick([]string{a, b, c}, Options{StickyKey: "task-1"})
		require.NoError(t, err)
		reordered := []string{c, b, a}
		second, err := p.Pick(reordered, Options{Strategy: Random, StickyKey: "task-1"})
		require.NoError(t, err)
		assert.Equal(t, []string{a, b, c}[first], reordered[second])
	})

	t.Run("sticky strategy hashes the key onto the ids", func(t *testing.T) {
		p := New()
		first, err := p.Pick([]string{a, b, c}, Options{Strategy: Sticky, StickyKey: "support-assistant"})
		require.NoError(t, err)
		second, err := p.Pick([]string{c, b, a}, Options{Strategy: Sticky, StickyKey: "support-assistant"})
		require.NoError(t, err)
		assert.Equal(t, []string{a, b, c}[first], []string{c, b, a}[second])
	})

	t.Run("sticky strategy without a key falls back to round robin", func(t *testing.T) {
		p := New()
		ids := []string{b, a}
		var picked []string
		for range 3 {
			got, err := p.Pick(ids, Options{Strategy: Sticky, Key: "invoice-extractor"})
			require.NoError(t, err)
			picked = append(picked, ids[got])
		}

		assert.Equal(t, []string{a, b, a}, picked)
	})

	t.Run("sticky follows remaining ids", func(t *testing.T) {
		p := New()
		all := []string{a, b, c}
		full, err := p.Pick(all, Options{StickyKey: "task-9"})
		require.NoError(t, err)

		var remaining []string
		for _, id := range all {
			if id != all[full] {
				remaining = append(remaining, id)
			}
		}

		got, err := p.Pick(remaining, Options{StickyKey: "task-9"})
		require.NoError(t, err)
		assert.NotEqual(t, all[full], remaining[got])
	})
}
