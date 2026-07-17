package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--mcp" {
		runMCPServer()
		return
	}

	report, regressed, err := runAudit()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(report)
	if regressed {
		os.Exit(1)
	}
}

func runAudit() (report string, regressed bool, err error) {
	cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "./...")
	out, execErr := cmd.CombinedOutput()
	if execErr != nil && len(out) == 0 {
		return "", false, fmt.Errorf("benchmark failed: %v", execErr)
	}

	results, parseErr := ParseBenchmarks(strings.NewReader(string(out)))
	if parseErr != nil {
		return "", false, fmt.Errorf("parse error: %v", parseErr)
	}
	if len(results) == 0 {
		return "", false, fmt.Errorf("no benchmarks found in output")
	}

	baseline, loadErr := LoadBaseline()
	if errors.Is(loadErr, os.ErrNotExist) {
		if saveErr := SaveBaseline(results); saveErr != nil {
			return "", false, fmt.Errorf("save baseline: %v", saveErr)
		}
		return buildNewBaselineReport(results), false, nil
	}
	if loadErr != nil {
		return "", false, fmt.Errorf("load baseline: %v", loadErr)
	}

	rows, hasRegression := Compare(baseline, results)
	return buildReport(rows, hasRegression), hasRegression, nil
}

func buildNewBaselineReport(results map[string]BenchmarkResult) string {
	var b strings.Builder
	b.WriteString("### 📊 Performance Audit Report\n\n")
	b.WriteString("No baseline found. Saved current metrics as new baseline.\n\n")
	b.WriteString("| Benchmark Name | Speed (ns/op) | Memory (B/op) | Allocs (allocs/op) |\n")
	b.WriteString("| :--- | ---: | ---: | ---: |\n")
	for name, r := range results {
		fmt.Fprintf(&b, "| `Benchmark%s` | %d ns | %d B | %d |\n", name, r.NsOp, r.BytesOp, r.AllocsOp)
	}
	return b.String()
}

func buildReport(rows []Row, hasRegression bool) string {
	var b strings.Builder
	b.WriteString("### 📊 Performance Audit Report\n\n")
	b.WriteString("| Benchmark Name | Metric | Baseline | Current | Delta | Status |\n")
	b.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")

	prevName := ""
	for _, r := range rows {
		displayName := ""
		if r.Name != prevName {
			displayName = fmt.Sprintf("`Benchmark%s`", r.Name)
			prevName = r.Name
		}
		status := "✅ Green"
		if r.Regressed {
			status = "❌ Regression"
		}
		sign := ""
		if r.DeltaPct > 0 {
			sign = "+"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s%.1f%% | %s |\n",
			displayName, r.Metric,
			fmtVal(r.Baseline, r.Metric),
			fmtVal(r.Current, r.Metric),
			sign, r.DeltaPct, status,
		)
	}

	b.WriteString("\n")
	if hasRegression {
		b.WriteString("⚠️ **Audit Result:** Performance regression detected in memory metrics. Optimization is strictly required before committing.")
	} else {
		b.WriteString("✅ **Audit Result:** All metrics within threshold. No regression detected.")
	}
	return b.String()
}

func fmtVal(v int64, metric string) string {
	switch {
	case strings.Contains(metric, "ns/op"):
		return fmt.Sprintf("%d ns", v)
	case strings.Contains(metric, "B/op"):
		return fmt.Sprintf("%d B", v)
	default:
		return fmt.Sprintf("%d", v)
	}
}
