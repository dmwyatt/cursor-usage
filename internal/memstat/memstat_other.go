//go:build !darwin

package memstat

import "fmt"

// Read is only implemented on macOS.
func Read() (*Stats, error) {
	return nil, fmt.Errorf("memory stats not supported on this platform")
}
