package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteFittingSectionsDropsTailWhenShort(t *testing.T) {
	sections := []string{
		"summary\nline2\n",
		"events\nline2\nline3\n",
		"memory\nline2\n",
	}

	var buf bytes.Buffer
	if err := writeFittingSections(&buf, 6, sections); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "summary") || !strings.Contains(got, "events") {
		t.Fatalf("expected summary and events, got:\n%s", got)
	}
	if strings.Contains(got, "memory") {
		t.Fatalf("expected memory to be dropped when short, got:\n%s", got)
	}
}

func TestWriteFittingSectionsKeepsAllWhenTall(t *testing.T) {
	sections := []string{"a\n", "b\n", "c\n"}
	var buf bytes.Buffer
	if err := writeFittingSections(&buf, 20, sections); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"a", "b", "c"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestWriteFittingSectionsUnknownBudgetKeepsAll(t *testing.T) {
	sections := []string{"a\n", "b\n", "c\n"}
	var buf bytes.Buffer
	if err := writeFittingSections(&buf, 0, sections); err != nil {
		t.Fatal(err)
	}
	if strings.Count(buf.String(), "\n") < 5 {
		t.Fatalf("expected all sections when budget unknown, got:\n%s", buf.String())
	}
}
