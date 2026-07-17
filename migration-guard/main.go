package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func run(dir string) (int, error) {
	pairs, err := ScanDirectory(dir)
	if err != nil {
		return 1, fmt.Errorf("error scanning directory: %w", err)
	}

	var results []AuditResult
	for _, pair := range pairs {
		var pairResults []AuditResult
		if pair.UpPath != "" {
			res, err := AnalyzeSQL(pair.UpPath)
			if err != nil {
				return 1, fmt.Errorf("error analyzing %s: %w", pair.UpPath, err)
			}
			pairResults = append(pairResults, res...)
		}

		// Down file rollback check
		hasDown := false
		if pair.DownPath != "" {
			fi, err := os.Stat(pair.DownPath)
			if err == nil && fi.Size() > 0 {
				hasDown = true
			}
		}

		if !hasDown {
			name := pair.BaseName + ".down.sql"
			pairResults = append(pairResults, AuditResult{
				File:      name,
				CheckItem: "Rollback Verification",
				Status:    StatusCritical,
				Details:   "Corresponding down migration file is missing or empty.",
			})
		}

		results = append(results, pairResults...)
	}

	if len(results) == 0 {
		fmt.Println("No migration files found.")
		return 0, nil
	}

	fmt.Println("### 🛡️ Database Migration Audit Report")
	fmt.Println()
	fmt.Println("| Migration File | Check Item | Status | Details / Remediation |")
	fmt.Println("| :--- | :--- | :--- | :--- |")

	var criticalCount int
	for _, res := range results {
		statusStr := "Green"
		if res.Status == StatusWarning {
			statusStr = "⚠️ Warning"
		} else if res.Status == StatusCritical {
			statusStr = "❌ Critical"
			criticalCount++
		}
		fmt.Printf("| `%s` | %s | %s | %s |\n", filepath.Base(res.File), res.CheckItem, statusStr, res.Details)
	}
	fmt.Println()

	if criticalCount > 0 {
		fmt.Println("❌ **Audit Result:** Database safety violations detected. You must generate the matching down/rollback SQL migration file before proceeding.")
		return 1, nil
	}

	fmt.Println("✅ **Audit Result:** All migrations passed safety checks.")
	return 0, nil
}

func main() {
	dir := flag.String("dir", "./migrations", "Directory containing migration files")
	flag.Parse()

	code, err := run(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}
