package output

import (
	"fmt"
	"strconv"

	"github.com/dmwyatt/cursor-usage/internal/api"
)

// TokenTotals is the sum of token usage across a set of events.
type TokenTotals struct {
	Input      int `json:"input"`
	Output     int `json:"output"`
	CacheRead  int `json:"cacheRead"`
	CacheWrite int `json:"cacheWrite"`
	Total      int `json:"total"`
}

// SumTokens adds input, output, cache-read, and cache-write tokens across events.
func SumTokens(events []api.UsageEvent) TokenTotals {
	var t TokenTotals
	for _, e := range events {
		t.Input += e.TokenUsage.InputTokens
		t.Output += e.TokenUsage.OutputTokens
		t.CacheRead += e.TokenUsage.CacheReadTokens
		t.CacheWrite += e.TokenUsage.CacheWriteTokens
	}
	t.Total = t.Input + t.Output + t.CacheRead + t.CacheWrite
	return t
}

// FormatTokenCount formats a count like the Cursor dashboard (140.8M, 217.4K).
func FormatTokenCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return strconv.Itoa(n)
	}
}

