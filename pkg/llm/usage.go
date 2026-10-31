// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"math"

	"github.com/samber/lo"
)

// USDToNano converts USD amounts into fixed-point nano-USD integers.
const USDToNano = 1_000_000_000

// Usage is token and cost usage reported by AI providers.
type Usage struct {
	PromptTokens int64
	TotalTokens  int64
	Cost         int64
}

// CostFromUSD converts an OpenRouter usage.cost value into fixed-point nano-USD.
func CostFromUSD(amount float64) int64 {
	return lo.Ternary(amount <= 0, int64(0), int64(math.Round(amount*USDToNano)))
}
