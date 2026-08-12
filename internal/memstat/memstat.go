package memstat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Stats is system memory usage in bytes, using Activity Monitor's breakdown.
type Stats struct {
	Physical   uint64 `json:"physical"`
	Used       uint64 `json:"used"`
	App        uint64 `json:"app"`
	Wired      uint64 `json:"wired"`
	Compressed uint64 `json:"compressed"`
	Cached     uint64 `json:"cached"`
	SwapUsed   uint64 `json:"swapUsed"`
	Pressure   string `json:"pressure,omitempty"`
}

var (
	pageSizeRe = regexp.MustCompile(`page size of (\d+) bytes`)
	statLineRe = regexp.MustCompile(`^([^:]+):\s+([\d.]+)`)
	swapUsedRe = regexp.MustCompile(`used\s*=\s*([\d.]+)([KMG])`)
)

// Compute builds Activity Monitor-style stats from page counts and sysctl values.
// App Memory = (pageable internal - purgeable) * page size
// Memory Used = App + Wired + Compressed
// Cached Files = (file-backed + purgeable) * page size
func Compute(physical, pageSize, pageableInternal, purgeable, wired, compressed, fileBacked, swapUsed uint64) Stats {
	if pageSize == 0 {
		pageSize = 4096
	}
	appPages := pageableInternal
	if pageableInternal > purgeable {
		appPages = pageableInternal - purgeable
	} else {
		appPages = 0
	}
	app := appPages * pageSize
	wiredBytes := wired * pageSize
	compressedBytes := compressed * pageSize
	return Stats{
		Physical:   physical,
		Used:       app + wiredBytes + compressedBytes,
		App:        app,
		Wired:      wiredBytes,
		Compressed: compressedBytes,
		Cached:     (fileBacked + purgeable) * pageSize,
		SwapUsed:   swapUsed,
	}
}

func parseVMStat(output string) (pageSize uint64, pages map[string]uint64, err error) {
	pages = make(map[string]uint64)
	m := pageSizeRe.FindStringSubmatch(output)
	if m == nil {
		return 0, nil, fmt.Errorf("parsing vm_stat page size")
	}
	pageSize, err = strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("parsing vm_stat page size: %w", err)
	}
	for _, line := range strings.Split(output, "\n") {
		sm := statLineRe.FindStringSubmatch(strings.TrimSpace(line))
		if sm == nil {
			continue
		}
		n, err := strconv.ParseUint(strings.TrimSuffix(sm[2], "."), 10, 64)
		if err != nil {
			continue
		}
		pages[strings.Trim(sm[1], `"`)] = n
	}
	return pageSize, pages, nil
}

func parseSwapUsed(output string) (uint64, error) {
	m := swapUsedRe.FindStringSubmatch(output)
	if m == nil {
		return 0, fmt.Errorf("parsing swap usage %q", output)
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parsing swap used: %w", err)
	}
	var mul float64
	switch m[2] {
	case "K":
		mul = 1024
	case "M":
		mul = 1024 * 1024
	case "G":
		mul = 1024 * 1024 * 1024
	}
	return uint64(val*mul + 0.5), nil
}

// FormatBytes formats a byte count like Activity Monitor (1024-based GB/MB).
func FormatBytes(b uint64) string {
	const (
		mb = 1024 * 1024
		gb = 1024 * 1024 * 1024
	)
	if b >= gb {
		return fmt.Sprintf("%.2f GB", float64(b)/float64(gb))
	}
	return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
}

// PressureFromLevel maps kern.memorystatus_vm_pressure_level to Activity Monitor colors.
// 1 = normal (green), 2 = warning (yellow), 4 = critical (red).
func PressureFromLevel(level uint64) string {
	switch level {
	case 2:
		return "yellow"
	case 4:
		return "red"
	default:
		return "green"
	}
}
