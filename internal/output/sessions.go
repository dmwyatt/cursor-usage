package output

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/jedib0t/go-pretty/v6/table"
)

// Session represents a group of events that occurred close together in time.
type Session struct {
	Start             time.Time           `json:"start"`
	End               time.Time           `json:"end"`
	Duration          time.Duration       `json:"duration"`
	Events            []api.UsageEvent    `json:"events"`
	TotalChargedCents float64             `json:"totalChargedCents"`
	ByModel           []SessionModelSummary `json:"byModel"`
}

// SessionModelSummary is a per-model breakdown within a session.
type SessionModelSummary struct {
	Model        string  `json:"model"`
	Events       int     `json:"events"`
	ChargedCents float64 `json:"chargedCents"`
}

type timedEvent struct {
	t     time.Time
	event api.UsageEvent
}

// GroupSessions groups events into sessions. Consecutive events less than gap
// apart belong to the same session. Events are sorted by timestamp internally.
func GroupSessions(events []api.UsageEvent, gap time.Duration) []Session {
	if len(events) == 0 {
		return nil
	}

	timed := make([]timedEvent, 0, len(events))
	for _, e := range events {
		n, err := strconv.ParseInt(e.Timestamp, 10, 64)
		if err != nil {
			continue
		}
		timed = append(timed, timedEvent{t: time.UnixMilli(n), event: e})
	}

	if len(timed) == 0 {
		return nil
	}

	sort.Slice(timed, func(i, j int) bool { return timed[i].t.Before(timed[j].t) })

	var sessions []Session
	current := []timedEvent{timed[0]}

	for _, te := range timed[1:] {
		last := current[len(current)-1]
		if te.t.Sub(last.t) > gap {
			sessions = append(sessions, buildSession(current))
			current = []timedEvent{te}
		} else {
			current = append(current, te)
		}
	}
	sessions = append(sessions, buildSession(current))

	return sessions
}

func buildSession(timed []timedEvent) Session {
	s := Session{
		Start:  timed[0].t,
		End:    timed[len(timed)-1].t,
		Events: make([]api.UsageEvent, len(timed)),
	}
	s.Duration = s.End.Sub(s.Start)

	models := map[string]*SessionModelSummary{}
	for i, te := range timed {
		s.Events[i] = te.event
		s.TotalChargedCents += te.event.ChargedCents

		m, ok := models[te.event.Model]
		if !ok {
			m = &SessionModelSummary{Model: te.event.Model}
			models[te.event.Model] = m
		}
		m.Events++
		m.ChargedCents += te.event.ChargedCents
	}

	for _, m := range models {
		s.ByModel = append(s.ByModel, *m)
	}
	sort.Slice(s.ByModel, func(i, j int) bool {
		return s.ByModel[i].ChargedCents > s.ByModel[j].ChargedCents
	})

	return s
}

// ActiveHours calculates total active hours from event timestamps.
func ActiveHours(timestamps []string, gap time.Duration) float64 {
	times := parseMsTimestamps(timestamps)
	if len(times) < 2 {
		return 0
	}

	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })

	var totalDuration time.Duration
	sessionStart := times[0]
	prev := times[0]

	for _, t := range times[1:] {
		if t.Sub(prev) > gap {
			totalDuration += prev.Sub(sessionStart)
			sessionStart = t
		}
		prev = t
	}
	totalDuration += prev.Sub(sessionStart)

	return totalDuration.Hours()
}

func parseMsTimestamps(timestamps []string) []time.Time {
	times := make([]time.Time, 0, len(timestamps))
	for _, ts := range timestamps {
		n, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			continue
		}
		times = append(times, time.UnixMilli(n))
	}
	return times
}

// RenderSessions writes a human-readable session table to w.
func RenderSessions(w io.Writer, sessions []Session) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)

	t.AppendHeader(table.Row{"Session", "Duration", "Events", "Cost"})

	var totalCents float64
	var totalEvents int
	var totalDuration time.Duration

	for i, s := range sessions {
		totalCents += s.TotalChargedCents
		totalEvents += len(s.Events)
		totalDuration += s.Duration

		if i > 0 {
			t.AppendSeparator()
		}

		sessionLabel := fmt.Sprintf("%s - %s",
			s.Start.Local().Format("2006-01-02 15:04"),
			s.End.Local().Format("15:04"),
		)
		t.AppendRow(table.Row{
			sessionLabel,
			formatDuration(s.Duration),
			len(s.Events),
			fmt.Sprintf("$%.2f", s.TotalChargedCents/100),
		})

		for _, m := range s.ByModel {
			t.AppendRow(table.Row{
				"  " + m.Model,
				"",
				m.Events,
				fmt.Sprintf("$%.2f", m.ChargedCents/100),
			})
		}
	}

	t.AppendSeparator()

	totalHours := totalDuration.Hours()
	footerLabel := fmt.Sprintf("Total (%d sessions)", len(sessions))
	if totalHours > 0 {
		footerLabel = fmt.Sprintf("Total (%d sessions, $%.2f/hr)",
			len(sessions), (totalCents/100)/totalHours)
	}

	t.AppendFooter(table.Row{
		footerLabel,
		formatDuration(totalDuration),
		totalEvents,
		fmt.Sprintf("$%.2f", totalCents/100),
	})

	t.Render()
	return nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
