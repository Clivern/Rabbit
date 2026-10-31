// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/llm"
	"github.com/clivern/sloth/pkg/stripe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripesdk "github.com/stripe/stripe-go/v87"
)

func TestUnitBilling(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}

	t.Run("status returns not found for a missing workspace", func(t *testing.T) {
		billing, _, _ := newBillingModule(nil, nil)

		_, err := billing.GetBillingStatus("ws-1")
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("status returns not found when the workspace has no subscription", func(t *testing.T) {
		billing, _, _ := newBillingModule(workspace, nil)

		_, err := billing.GetBillingStatus("ws-1")
		assert.ErrorIs(t, err, ErrBillingSubscriptionNotFound)
	})

	t.Run("status returns the balance and the token rate", func(t *testing.T) {
		customer := "cus_123"
		billing, _, _ := newBillingModule(workspace, &db.Subscription{
			Id:                 "sub-1",
			WorkspaceId:        "ws-1",
			ProviderCustomerId: &customer,
			AITokensBalance:    900,
			CreatedAt:          time.Now(),
		})

		status, err := billing.GetBillingStatus("ws-1")
		require.NoError(t, err)
		assert.Equal(t, int64(900), status.AITokensBalance)
		assert.Equal(t, "cus_123", status.ProviderCustomerId)
		assert.Equal(t, stripe.TokensPerUSD(), status.TokensPerUsd)
	})

	t.Run(
		"usage adds the used tokens to the balance and reports the member limit",
		func(t *testing.T) {
			billing, _, _ := newBillingModule(
				workspace,
				&db.Subscription{Id: "sub-1", WorkspaceId: "ws-1", AITokensBalance: 1000},
			)
			deps := UsageSnapshotDeps{
				WorkspaceUserRepository: &fakeWorkspaceUsers{total: 2},
				UsageRepository: &fakeUsage{
					quantities: map[string]int64{
						db.UsageTypeAITokens: 300,
						db.UsageTypeAICost:   llm.USDToNano,
					},
				},
			}

			usage, err := billing.GetBillingUsage("ws-1", deps)
			require.NoError(t, err)
			assert.Equal(t, int64(2), usage.Used.WorkspaceMembers)
			assert.Equal(t, int64(3), usage.Limits.WorkspaceMembers)
			assert.Equal(t, int64(1300), usage.Limits.AITokens)
			assert.NotEmpty(t, usage.PeriodReset)
		},
	)

	t.Run("usage returns the usage repository error", func(t *testing.T) {
		billing, _, _ := newBillingModule(workspace, &db.Subscription{Id: "sub-1", WorkspaceId: "ws-1"})
		deps := UsageSnapshotDeps{
			WorkspaceUserRepository: &fakeWorkspaceUsers{},
			UsageRepository:         &fakeUsage{err: errors.New("boom")},
		}

		_, err := billing.GetBillingUsage("ws-1", deps)
		assert.EqualError(t, err, "boom")
	})

	t.Run("workspace subscription returns not found when missing", func(t *testing.T) {
		billing, _, _ := newBillingModule(workspace, nil)

		_, err := billing.GetWorkspaceSubscription("ws-1")
		assert.ErrorIs(t, err, ErrBillingSubscriptionNotFound)
	})

	t.Run("workspace subscription returns the stored subscription", func(t *testing.T) {
		billing, _, _ := newBillingModule(workspace, &db.Subscription{Id: "sub-1"})

		subscription, err := billing.GetWorkspaceSubscription("ws-1")
		require.NoError(t, err)
		assert.Equal(t, db.Id("sub-1"), subscription.Id)
	})

	t.Run(
		"checkout reports billing as disabled when stripe is not configured",
		func(t *testing.T) {
			billing, _, _ := newBillingModule(workspace, &db.Subscription{Id: "sub-1"})

			_, err := billing.CreateCheckoutSession(
				context.Background(),
				"ws-1",
				&db.User{Email: "ada@example.com"},
				&BillingCheckoutRequest{AmountCents: 1000},
				"ok",
				"cancel",
			)
			assert.ErrorIs(t, err, stripe.ErrBillingDisabled)
		},
	)

	t.Run("checkout rejects amounts below the minimum purchase", func(t *testing.T) {
		enableBilling(t)
		billing, _, _ := newBillingModule(workspace, &db.Subscription{Id: "sub-1"})

		_, err := billing.CreateCheckoutSession(
			context.Background(),
			"ws-1",
			&db.User{Email: "ada@example.com"},
			&BillingCheckoutRequest{AmountCents: stripe.MinTokenPurchaseCents - 1},
			"ok",
			"cancel",
		)
		assert.ErrorIs(t, err, stripe.ErrInvalidAmount)
	})

	t.Run("checkout returns not found when the workspace has no subscription", func(t *testing.T) {
		enableBilling(t)
		billing, _, _ := newBillingModule(workspace, nil)

		_, err := billing.CreateCheckoutSession(
			context.Background(),
			"ws-1",
			&db.User{Email: "ada@example.com"},
			&BillingCheckoutRequest{AmountCents: 1000},
			"ok",
			"cancel",
		)
		assert.ErrorIs(t, err, ErrBillingSubscriptionNotFound)
	})

	t.Run("portal reports billing as disabled when stripe is not configured", func(t *testing.T) {
		billing, _, _ := newBillingModule(workspace, &db.Subscription{Id: "sub-1"})

		_, err := billing.CreatePortalSession(context.Background(), "ws-1", "return")
		assert.ErrorIs(t, err, stripe.ErrBillingDisabled)
	})

	t.Run("portal is unavailable until the workspace has a stripe customer", func(t *testing.T) {
		enableBilling(t)
		billing, _, _ := newBillingModule(workspace, &db.Subscription{Id: "sub-1"})

		_, err := billing.CreatePortalSession(context.Background(), "ws-1", "return")
		assert.ErrorIs(t, err, ErrBillingPortalUnavailable)
	})

	t.Run("webhook reports billing as disabled when stripe is not configured", func(t *testing.T) {
		billing, _, _ := newBillingModule(workspace, nil)

		err := billing.HandleWebhook([]byte(`{}`), "bad")
		assert.ErrorIs(t, err, stripe.ErrBillingDisabled)
	})

	t.Run("webhook rejects a signature that does not match", func(t *testing.T) {
		enableBilling(t)
		billing, _, _ := newBillingModule(workspace, nil)

		err := billing.HandleWebhook([]byte(`{"type":"checkout.session.completed"}`), "t=1,v1=bad")
		assert.Error(t, err)
	})

	t.Run("credit ignores purchases without a workspace", func(t *testing.T) {
		client := enableBilling(t)
		billing, _, purchases := newBillingModule(workspace, nil)

		err := billing.CreditTokenPurchase(client, &stripesdk.CheckoutSession{ID: "cs_1", AmountTotal: 1000})
		require.NoError(t, err)
		assert.Empty(t, purchases.created)
	})

	t.Run("credit ignores purchases that convert to no tokens", func(t *testing.T) {
		client := enableBilling(t)
		billing, _, purchases := newBillingModule(workspace, nil)

		err := billing.CreditTokenPurchase(client, &stripesdk.CheckoutSession{
			ID:          "cs_1",
			AmountTotal: 0,
			Metadata:    map[string]string{"workspaceId": "ws-1"},
		})
		require.NoError(t, err)
		assert.Empty(t, purchases.created)
	})

	t.Run("credit skips a session that was already recorded", func(t *testing.T) {
		client := enableBilling(t)
		billing, subscriptions, purchases := newBillingModule(workspace, nil)
		purchases.inserted = false

		err := billing.CreditTokenPurchase(client, &stripesdk.CheckoutSession{
			ID:          "cs_1",
			AmountTotal: 1000,
			Metadata:    map[string]string{"workspaceId": "ws-1"},
		})
		require.NoError(t, err)
		assert.Empty(t, subscriptions.added)
	})

	t.Run(
		"credit records the purchase, adds the tokens and stores the customer",
		func(t *testing.T) {
			client := enableBilling(t)
			subscription := &db.Subscription{Id: "sub-1", WorkspaceId: "ws-1"}
			billing, subscriptions, purchases := newBillingModule(workspace, subscription)
			tokens := client.Config().TokensForCents(1000)

			err := billing.CreditTokenPurchase(client, &stripesdk.CheckoutSession{
				ID:          "cs_1",
				AmountTotal: 1000,
				Metadata:    map[string]string{"workspaceId": "ws-1"},
				Customer:    &stripesdk.Customer{ID: "cus_123"},
			})
			require.NoError(t, err)

			require.Len(t, purchases.created, 1)
			assert.Equal(t, tokens, purchases.created[0].Tokens)
			assert.Equal(t, []int64{tokens}, subscriptions.added)
			require.Len(t, subscriptions.updated, 1)
			require.NotNil(t, subscriptions.updated[0].ProviderCustomerId)
			assert.Equal(t, "cus_123", *subscriptions.updated[0].ProviderCustomerId)
		},
	)

	t.Run("credit wraps the purchase repository error", func(t *testing.T) {
		client := enableBilling(t)
		billing, _, purchases := newBillingModule(workspace, nil)
		purchases.err = errors.New("boom")

		err := billing.CreditTokenPurchase(client, &stripesdk.CheckoutSession{
			ID:          "cs_1",
			AmountTotal: 1000,
			Metadata:    map[string]string{"workspaceId": "ws-1"},
		})
		assert.ErrorContains(t, err, "record token purchase")
	})
}
