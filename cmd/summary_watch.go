package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/dmwyatt/cursor-usage/internal/memstat"
	"github.com/dmwyatt/cursor-usage/internal/output"
)

const (
	ansiEnterAlt = "\033[?1049h"
	ansiLeaveAlt = "\033[?1049l"
	ansiHideCur  = "\033[?25l"
	ansiShowCur  = "\033[?25h"
	ansiHomeClr  = "\033[H\033[2J"
)

// runSummaryWatch refreshes the summary in the alternate screen. Prefer this
// over watch(1): each tick resets the cursor to home so phone scroll offsets
// cannot stack garbage into the next frame, and content is clipped to the
// live tty size so there is nothing that needs scrolling.
func runSummaryWatch(interval time.Duration) error {
	if interval < time.Second {
		return fmt.Errorf("--interval must be at least 1s")
	}

	out := os.Stdout
	if _, err := fmt.Fprint(out, ansiEnterAlt+ansiHideCur); err != nil {
		return err
	}
	defer fmt.Fprint(out, ansiShowCur+ansiLeaveAlt)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	render := func() {
		fmt.Fprint(out, ansiHomeClr)
		height := output.OutputHeight()
		// Leave one row for the status line under the dashboard.
		budget := height
		if budget > 1 {
			budget--
		}

		summary, tokens, recent, mem, err := fetchSummaryDashboard()
		if err != nil {
			fmt.Fprintf(out, "error: %v\n\nCtrl+C to quit · retrying every %s\n", err, interval)
			return
		}
		_ = output.RenderSummaryScreenBudget(out, budget, summary, tokens, recent, mem)
		fmt.Fprintf(out, "\nrefresh every %s · Ctrl+C quit", interval)
	}

	render()
	for {
		select {
		case <-interrupt:
			return nil
		case <-winch:
			render()
		case <-ticker.C:
			render()
		}
	}
}

func fetchSummaryDashboard() (*api.UsageSummary, *output.TokenTotals, []api.UsageEvent, *memstat.Stats, error) {
	summary, err := apiClient.GetUsageSummary()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	startMs, err := billingCycleStartMillis(summary.BillingCycleStart)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	eventsResp, err := apiClient.GetAllFilteredUsageEvents(api.EventsRequest{
		StartDate: startMs,
		PageSize:  200,
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tokens := output.SumTokens(eventsResp.UsageEventsDisplay)
	recent := eventsResp.UsageEventsDisplay
	if len(recent) > recentEventCount {
		recent = recent[:recentEventCount]
	}
	mem, _ := memstat.Read()
	return summary, &tokens, recent, mem, nil
}
