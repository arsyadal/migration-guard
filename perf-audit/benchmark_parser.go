package main

import (
	"bufio"
	"io"
	"math"
	"regexp"
	"strconv"
)

// ns/op can be a float (e.g. 0.3172 ns/op) so we match \d+(?:\.\d+)?
var benchRegex = regexp.MustCompile(`^Benchmark([a-zA-Z0-9_]+)-\d+\s+\d+\s+(\d+(?:\.\d+)?)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)

type BenchmarkResult struct {
	NsOp     int64 `json:"ns_op"`
	BytesOp  int64 `json:"bytes_op"`
	AllocsOp int64 `json:"allocs_op"`
}

func ParseBenchmarks(r io.Reader) (map[string]BenchmarkResult, error) {
	results := make(map[string]BenchmarkResult)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		m := benchRegex.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		nsF, _ := strconv.ParseFloat(m[2], 64)
		bytesOp, _ := strconv.ParseInt(m[3], 10, 64)
		allocsOp, _ := strconv.ParseInt(m[4], 10, 64)
		results[m[1]] = BenchmarkResult{
			NsOp:     int64(math.Round(nsF)),
			BytesOp:  bytesOp,
			AllocsOp: allocsOp,
		}
	}
	return results, scanner.Err()
}
