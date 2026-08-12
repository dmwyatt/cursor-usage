package memstat

import "testing"

const sampleVMStat = `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                                     5241.
Pages active:                                 114968.
Pages inactive:                               114895.
Pages speculative:                               527.
Pages throttled:                                   0.
Pages wired down:                             119674.
Pages purgeable:                                2015.
File-backed pages:                             72639.
Anonymous pages:                              157751.
Pages stored in compressor:                   362549.
Pages occupied by compressor:                 133050.
`

func TestParseVMStat(t *testing.T) {
	pageSize, pages, err := parseVMStat(sampleVMStat)
	if err != nil {
		t.Fatal(err)
	}
	if pageSize != 16384 {
		t.Errorf("page size: got %d", pageSize)
	}
	if pages["Pages wired down"] != 119674 {
		t.Errorf("wired: got %d", pages["Pages wired down"])
	}
	if pages["Pages occupied by compressor"] != 133050 {
		t.Errorf("compressed: got %d", pages["Pages occupied by compressor"])
	}
	if pages["File-backed pages"] != 72639 {
		t.Errorf("file-backed: got %d", pages["File-backed pages"])
	}
}

func TestParseSwapUsed(t *testing.T) {
	n, err := parseSwapUsed("total = 2048.00M  used = 709.1M  free = 1338.90M  (encrypted)")
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatBytes(n); got != "709.1 MB" {
		t.Errorf("got %q", got)
	}
}

func TestCompute(t *testing.T) {
	s := Compute(8589934592, 16384, 143651, 2015, 119674, 133050, 72639, 743440384)
	if s.Physical != 8589934592 {
		t.Errorf("physical: got %d", s.Physical)
	}
	if s.Used != s.App+s.Wired+s.Compressed {
		t.Errorf("used %d != app+wired+compressed %d", s.Used, s.App+s.Wired+s.Compressed)
	}
	if s.App == 0 || s.Wired == 0 || s.Compressed == 0 {
		t.Errorf("expected non-zero breakdown: %+v", s)
	}
	if s.Cached == 0 {
		t.Error("expected non-zero cached files")
	}
}

func TestFormatBytes(t *testing.T) {
	if got := FormatBytes(8589934592); got != "8.00 GB" {
		t.Errorf("physical: got %q", got)
	}
	if got := FormatBytes(512 * 1024 * 1024); got != "512.0 MB" {
		t.Errorf("swap: got %q", got)
	}
}

func TestPressureFromLevel(t *testing.T) {
	cases := map[uint64]string{0: "green", 1: "green", 2: "yellow", 4: "red"}
	for level, want := range cases {
		if got := PressureFromLevel(level); got != want {
			t.Errorf("level %d: got %q, want %q", level, got, want)
		}
	}
}

func TestRead(t *testing.T) {
	s, err := Read()
	if err != nil {
		t.Skip(err)
	}
	if s.Physical == 0 {
		t.Fatal("physical memory is 0")
	}
	if s.Used != s.App+s.Wired+s.Compressed {
		t.Errorf("used %d != app+wired+compressed", s.Used)
	}
	switch s.Pressure {
	case "green", "yellow", "red":
	default:
		t.Errorf("unexpected pressure %q", s.Pressure)
	}
}
