package dateparse

import (
	"strconv"
	"testing"
	"time"
)

var testNow = time.Date(2026, 4, 5, 14, 30, 0, 0, time.UTC)

func TestToMillisISODate(t *testing.T) {
	ms, err := ToMillis("2026-04-01", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if ms != strconv.FormatInt(expected.UnixMilli(), 10) {
		t.Errorf("expected %d, got %s", expected.UnixMilli(), ms)
	}
}

func TestToMillisISODatetime(t *testing.T) {
	ms, err := ToMillis("2026-04-01T14:30:00", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2026, 4, 1, 14, 30, 0, 0, time.UTC)
	if ms != strconv.FormatInt(expected.UnixMilli(), 10) {
		t.Errorf("expected %d, got %s", expected.UnixMilli(), ms)
	}
}

func TestToMillisRelativeDays(t *testing.T) {
	ms, err := ToMillis("7d", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC)
	if ms != strconv.FormatInt(expected.UnixMilli(), 10) {
		t.Errorf("expected %d, got %s", expected.UnixMilli(), ms)
	}
}

func TestToMillisToday(t *testing.T) {
	ms, err := ToMillis("today", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	if ms != strconv.FormatInt(expected.UnixMilli(), 10) {
		t.Errorf("expected %d, got %s", expected.UnixMilli(), ms)
	}
}

func TestToMillisYesterday(t *testing.T) {
	ms, err := ToMillis("yesterday", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)
	if ms != strconv.FormatInt(expected.UnixMilli(), 10) {
		t.Errorf("expected %d, got %s", expected.UnixMilli(), ms)
	}
}

func TestToMillisInvalidFormat(t *testing.T) {
	_, err := ToMillis("not-a-date", testNow)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestEndOfDayMillis(t *testing.T) {
	ms, err := EndOfDayMillis("2026-04-01", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n, _ := strconv.ParseInt(ms, 10, 64)
	got := time.UnixMilli(n).UTC()

	if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
		t.Errorf("expected end of day, got %v", got)
	}
}
