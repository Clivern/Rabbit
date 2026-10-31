// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/llm"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
)

// Usage provides workspace consumption helpers.
type Usage struct{}

// WorkspaceUsageMetrics is current workspace consumption for billing.
type WorkspaceUsageMetrics struct {
	WorkspaceMembers int64   `json:"workspaceMembers"`
	AITokens         int64   `json:"aiTokens"`
	AICost           float64 `json:"aiCost"`
}

// UsageSnapshotDeps holds repositories needed to load workspace usage.
type UsageSnapshotDeps struct {
	WorkspaceUserRepository db.WorkspaceUserRepository
	UsageRepository         db.UsageRepository
}

// NewUsage returns a usage module.
func NewUsage() *Usage {
	return &Usage{}
}

// MembersCount returns the number of members in a workspace.
func (u *Usage) MembersCount(workspaceUsers db.WorkspaceUserRepository, workspaceId db.Id) (int64, error) {
	return workspaceUsers.CountByWorkspaceId(workspaceId)
}

// GetWorkspaceUsage returns all billing usage metrics for a workspace.
func (u *Usage) GetWorkspaceUsage(deps UsageSnapshotDeps, workspaceId db.Id) (*WorkspaceUsageMetrics, error) {
	members, err := u.MembersCount(deps.WorkspaceUserRepository, workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count workspace members")
		return nil, err
	}

	aiTokens, err := u.AITokens(deps.UsageRepository, workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get AI token usage")
		return nil, err
	}

	aiCost, err := u.AICost(deps.UsageRepository, workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get AI cost usage")
		return nil, err
	}

	return &WorkspaceUsageMetrics{
		WorkspaceMembers: members,
		AITokens:         aiTokens,
		AICost:           aiCost,
	}, nil
}

// AITokens returns AI token usage for the current calendar month.
func (u *Usage) AITokens(usage db.UsageRepository, workspaceId db.Id) (int64, error) {
	start, _ := util.CurrentMonthPeriod()

	return usage.GetQuantityByPeriod(workspaceId, db.UsageTypeAITokens, start)
}

// AICost returns AI spend in USD for the current calendar month.
func (u *Usage) AICost(usage db.UsageRepository, workspaceId db.Id) (float64, error) {
	start, _ := util.CurrentMonthPeriod()
	cost, err := usage.GetQuantityByPeriod(workspaceId, db.UsageTypeAICost, start)
	if err != nil {
		return 0, err
	}

	return float64(cost) / llm.USDToNano, nil
}

// IncrementAIUsage adds AI token and cost usage for the current calendar month. TODO: remove if not used in future.
func (u *Usage) IncrementAIUsage(usage db.UsageRepository, subscriptions db.SubscriptionRepository, workspaceId db.Id, tokens, cost int64) error {
	if tokens == 0 && cost == 0 {
		return nil
	}

	start, end := util.CurrentMonthPeriod()

	if tokens != 0 {
		if err := usage.IncrementByPeriod(
			workspaceId,
			db.UsageTypeAITokens,
			start,
			end,
			tokens,
			db.UsageUnitTokens,
		); err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Msg("Failed to increment AI usage")
			return err
		}

		if err := subscriptions.ConsumeTokens(workspaceId, tokens); err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Int64("tokens", tokens).
				Msg("Failed to consume AI tokens")
			return err
		}
	}

	if cost != 0 {
		if err := usage.IncrementByPeriod(
			workspaceId,
			db.UsageTypeAICost,
			start,
			end,
			cost,
			db.UsageUnitNanoUSD,
		); err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Msg("Failed to increment AI usage")
			return err
		}
	}

	log.Debug().
		Str("workspaceId", workspaceId.String()).
		Int64("tokens", tokens).
		Int64("cost", cost).
		Msg("AI usage incremented")

	return nil
}
