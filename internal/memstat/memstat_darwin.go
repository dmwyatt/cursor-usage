//go:build darwin

package memstat

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Read returns current system memory stats matching Activity Monitor.
func Read() (*Stats, error) {
	vmOut, err := exec.Command("/usr/bin/vm_stat").Output()
	if err != nil {
		return nil, fmt.Errorf("running vm_stat: %w", err)
	}
	pageSize, pages, err := parseVMStat(string(vmOut))
	if err != nil {
		return nil, err
	}

	physical, err := sysctlUint("hw.memsize")
	if err != nil {
		return nil, err
	}
	internal, err := sysctlUint("vm.page_pageable_internal_count")
	if err != nil {
		return nil, err
	}
	swapOut, err := exec.Command("/usr/sbin/sysctl", "-n", "vm.swapusage").Output()
	if err != nil {
		return nil, fmt.Errorf("reading vm.swapusage: %w", err)
	}
	swapUsed, err := parseSwapUsed(string(swapOut))
	if err != nil {
		return nil, err
	}

	stats := Compute(
		physical,
		pageSize,
		internal,
		pages["Pages purgeable"],
		pages["Pages wired down"],
		pages["Pages occupied by compressor"],
		pages["File-backed pages"],
		swapUsed,
	)
	if level, err := sysctlUint("kern.memorystatus_vm_pressure_level"); err == nil {
		stats.Pressure = PressureFromLevel(level)
	}
	return &stats, nil
}

func sysctlUint(name string) (uint64, error) {
	out, err := exec.Command("/usr/sbin/sysctl", "-n", name).Output()
	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", name, err)
	}
	n, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", name, err)
	}
	return n, nil
}
