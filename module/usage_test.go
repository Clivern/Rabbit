// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/llm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitUsage(t *testing.T) {
	t.Run("members count comes from the membership repository", func(t *testing.T) {
		count, err := NewUsage().MembersCount(&fakeWorkspaceUsers{total: 4}, "ws-1")
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})

	t.Run("ai cost is converted from nano dollars to dollars", func(t *testing.T) {
		usage := &fakeUsage{quantities: map[string]int64{db.UsageTypeAICost: 2 * llm.USDToNano}}

		cost, err := NewUsage().AICost(usage, "ws-1")
		require.NoError(t, err)
		assert.InDelta(t, 2.0, cost, 1e-9)
	})

	t.Run("ai tokens and cost return repository errors", func(t *testing.T) {
		_, err := NewUsage().AITokens(&fakeUsage{err: errors.New("boom")}, "ws-1")
		assert.EqualError(t, err, "boom")

		_, err = NewUsage().AICost(&fakeUsage{err: errors.New("boom")}, "ws-1")
		assert.EqualError(t, err, "boom")
	})

	t.Run("workspace usage combines members, tokens and cost", func(t *testing.T) {
		deps := UsageSnapshotDeps{
			WorkspaceUserRepository: &fakeWorkspaceUsers{total: 3},
			UsageRepository: &fakeUsage{quantities: map[string]int64{
				db.UsageTypeAITokens: 1200,
				db.UsageTypeAICost:   llm.USDToNano,
			}},
		}

		metrics, err := NewUsage().GetWorkspaceUsage(deps, "ws-1")
		require.NoError(t, err)
		assert.Equal(t, int64(3), metrics.WorkspaceMembers)
		assert.Equal(t, int64(1200), metrics.AITokens)
		assert.InDelta(t, 1.0, metrics.AICost, 1e-9)
	})

	t.Run("workspace usage returns the first repository error", func(t *testing.T) {
		deps := UsageSnapshotDeps{
			WorkspaceUserRepository: &fakeWorkspaceUsers{err: errors.New("boom")},
			UsageRepository:         &fakeUsage{},
		}

		_, err := NewUsage().GetWorkspaceUsage(deps, "ws-1")
		assert.EqualError(t, err, "boom")
	})

	t.Run("increment does nothing when tokens and cost are zero", func(t *testing.T) {
		usage := &fakeUsage{}
		subscriptions := &fakeSubscriptions{}

		require.NoError(t, NewUsage().IncrementAIUsage(usage, subscriptions, "ws-1", 0, 0))
		assert.Empty(t, usage.incremented)
		assert.Empty(t, subscriptions.consumed)
	})

	t.Run("increment records tokens and consumes the subscription balance", func(t *testing.T) {
		usage := &fakeUsage{}
		subscriptions := &fakeSubscriptions{}

		require.NoError(t, NewUsage().IncrementAIUsage(usage, subscriptions, "ws-1", 500, 0))
		assert.Equal(t, []int64{500}, usage.incremented)
		assert.Equal(t, []int64{500}, subscriptions.consumed)
	})

	t.Run("increment records cost without consuming tokens", func(t *testing.T) {
		usage := &fakeUsage{}
		subscriptions := &fakeSubscriptions{}

		require.NoError(t, NewUsage().IncrementAIUsage(usage, subscriptions, "ws-1", 0, 75))
		assert.Equal(t, []int64{75}, usage.incremented)
		assert.Empty(t, subscriptions.consumed)
	})

	t.Run("increment returns the usage error and skips the consume", func(t *testing.T) {
		subscriptions := &fakeSubscriptions{}

		err := NewUsage().
			IncrementAIUsage(&fakeUsage{err: errors.New("boom")}, subscriptions, "ws-1", 500, 0)
		assert.EqualError(t, err, "boom")
		assert.Empty(t, subscriptions.consumed)
	})
}
