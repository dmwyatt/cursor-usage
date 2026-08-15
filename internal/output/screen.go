package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/dmwyatt/cursor-usage/internal/memstat"
)

// RenderSummaryScreen writes the summary dashboard. Recent events are preferred
// over memory when the terminal is too short to show everything — refresh UIs
// (watch(1) or --watch) cannot scroll clipped rows.
func RenderSummaryScreen(w io.Writer, s *api.UsageSummary, tokens *TokenTotals, recent []api.UsageEvent, mem *memstat.Stats) error {
	return RenderSummaryScreenBudget(w, OutputHeight(), s, tokens, recent, mem)
}

// RenderSummaryScreenBudget is like RenderSummaryScreen but uses an explicit
// height budget (0 means unlimited / unknown).
func RenderSummaryScreenBudget(w io.Writer, height int, s *api.UsageSummary, tokens *TokenTotals, recent []api.UsageEvent, mem *memstat.Stats) error {
	var summaryBuf, eventsBuf, memoryBuf bytes.Buffer

	if err := RenderSummary(&summaryBuf, s, tokens); err != nil {
		return err
	}
	if len(recent) > 0 {
		if err := RenderRecentEvents(&eventsBuf, recent); err != nil {
			return err
		}
	}
	if mem != nil {
		if err := RenderMemory(&memoryBuf, mem); err != nil {
			return err
		}
	}

	sections := []string{summaryBuf.String()}
	if eventsBuf.Len() > 0 {
		sections = append(sections, eventsBuf.String())
	}
	if memoryBuf.Len() > 0 {
		sections = append(sections, memoryBuf.String())
	}

	return writeFittingSections(w, height, sections)
}

func writeFittingSections(w io.Writer, budget int, sections []string) error {
	used := 0
	for _, section := range sections {
		if section == "" {
			continue
		}
		lines := strings.Count(section, "\n")
		need := lines
		if used > 0 {
			need++ // blank separator
		}
		if budget > 0 && used > 0 && used+need > budget {
			// Drop this and all lower-priority sections (events beat memory).
			return nil
		}
		if used > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, section); err != nil {
			return err
		}
		used += need
	}
	return nil
}
