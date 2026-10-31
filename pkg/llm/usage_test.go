// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitCostFromUSD(t *testing.T) {
	t.Run("positive amount", func(t *testing.T) {
		assert.Equal(t, int64(1_500_000_000), CostFromUSD(1.5))
	})

	t.Run("zero and negative", func(t *testing.T) {
		assert.Equal(t, int64(0), CostFromUSD(0))
		assert.Equal(t, int64(0), CostFromUSD(-1))
	})

	t.Run("rounds", func(t *testing.T) {
		assert.Equal(t, int64(0), CostFromUSD(0.0000000004))
		assert.Equal(t, int64(1), CostFromUSD(0.0000000005))
	})
}
