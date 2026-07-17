package main

import (
	"encoding/json"
	"os"
	"time"
)

const baselineFile = ".perf-baseline.json"
const regressionThreshold = 0.10

type Baseline struct {
	Timestamp  string                     `json:"timestamp"`
	Benchmarks map[string]BenchmarkResult `json:"benchmarks"`
}

type Row struct {
	Name      string
	Metric    string
	Baseline  int64
	Current   int64
	DeltaPct  float64
	Regressed bool
}

func LoadBaseline() (*Baseline, error) {
	data, err := os.ReadFile(baselineFile)
	if err != nil {
		return nil, err
	}
	var b Baseline
	return &b, json.Unmarshal(data, &b)
}

func SaveBaseline(results map[string]BenchmarkResult) error {
	b := Baseline{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Benchmarks: results,
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(baselineFile, data, 0644)
}

func Compare(baseline *Baseline, current map[string]BenchmarkResult) ([]Row, bool) {
	var rows []Row
	hasRegression := false
	for name, cur := range current {
		base, ok := baseline.Benchmarks[name]
		if !ok {
			continue
		}
		rows = append(rows, row(name, "Speed (`ns/op`)", base.NsOp, cur.NsOp, false))
		memRow := row(name, "Memory (`B/op`)", base.BytesOp, cur.BytesOp, true)
		allocRow := row(name, "Allocs (`allocs/op`)", base.AllocsOp, cur.AllocsOp, true)
		if memRow.Regressed || allocRow.Regressed {
			hasRegression = true
		}
		rows = append(rows, memRow, allocRow)
	}
	return rows, hasRegression
}

func row(name, metric string, base, cur int64, checkReg bool) Row {
	var delta float64
	if base != 0 {
		delta = float64(cur-base) / float64(base)
	}
	return Row{
		Name:      name,
		Metric:    metric,
		Baseline:  base,
		Current:   cur,
		DeltaPct:  delta * 100,
		Regressed: checkReg && delta > regressionThreshold,
	}
}
