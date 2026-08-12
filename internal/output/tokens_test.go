package output

import (
	"testing"

	"github.com/dmwyatt/cursor-usage/internal/api"
)

func TestSumTokens(t *testing.T) {
	got := SumTokens([]api.UsageEvent{
		{TokenUsage: api.TokenUsage{InputTokens: 1000, OutputTokens: 200, CacheReadTokens: 8000, CacheWriteTokens: 50}},
		{TokenUsage: api.TokenUsage{InputTokens: 500, OutputTokens: 100, CacheReadTokens: 2000, CacheWriteTokens: 0}},
	})
	if got.Input != 1500 || got.Output != 300 || got.CacheRead != 10000 || got.CacheWrite != 50 || got.Total != 11850 {
		t.Errorf("unexpected totals: %+v", got)
	}
}

func TestFormatTokenCount(t *testing.T) {
	cases := map[int]string{
		140_800_000: "140.8M",
		217_400:     "217.4K",
		999:         "999",
		0:           "0",
	}
	for n, want := range cases {
		if got := FormatTokenCount(n); got != want {
			t.Errorf("%d: got %q, want %q", n, got, want)
		}
	}
}
