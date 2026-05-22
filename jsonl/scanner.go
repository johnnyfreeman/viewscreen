// Package jsonl provides shared JSONL stream scanning helpers.
package jsonl

import (
	"bufio"
	"io"
)

// MaxScannerCapacity is the maximum JSONL line size accepted by viewscreen.
const MaxScannerCapacity = 10 * 1024 * 1024

// NewScanner creates a scanner configured for large agent JSONL events.
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, MaxScannerCapacity)
	scanner.Buffer(buf, MaxScannerCapacity)
	return scanner
}
