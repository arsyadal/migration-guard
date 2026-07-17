package main

import (
	"os"
	"regexp"
)

type AuditStatus string

const (
	StatusGreen    AuditStatus = "Green"
	StatusWarning  AuditStatus = "Warning"
	StatusCritical AuditStatus = "Critical"
)

type AuditResult struct {
	File      string
	CheckItem string
	Status    AuditStatus
	Details   string
}

func stripComments(sql string) string {
	// Strip multi-line comments
	reMulti := regexp.MustCompile(`(?s)/\*\*.*?\*/|/\*.*?\*/`)
	sql = reMulti.ReplaceAllString(sql, "")
	// Strip single-line comments
	reSingle := regexp.MustCompile(`--.*`)
	sql = reSingle.ReplaceAllString(sql, "")
	return sql
}

func AnalyzeSQL(filePath string) ([]AuditResult, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	content := stripComments(string(contentBytes))
	var results []AuditResult

	// Rule 1: RENAME COLUMN (Critical)
	reRename := regexp.MustCompile(`(?i)alter\s+table\s+.*\s+rename\s+(column\s+)?`)
	if reRename.MatchString(content) {
		results = append(results, AuditResult{
			File:      filePath,
			CheckItem: "Rename Column",
			Status:    StatusCritical,
			Details:   "ALTER TABLE RENAME COLUMN breaks compatibility.",
		})
	}

	// Rule 2: DROP TABLE or DROP COLUMN (Critical)
	reDrop := regexp.MustCompile(`(?i)drop\s+(table|column)\s+`)
	if reDrop.MatchString(content) {
		results = append(results, AuditResult{
			File:      filePath,
			CheckItem: "Destructive Commands",
			Status:    StatusCritical,
			Details:   "DROP TABLE or DROP COLUMN operation detected.",
		})
	}

	// Rule 3: ADD COLUMN NOT NULL without DEFAULT (Warning)
	reAddNotNull := regexp.MustCompile(`(?i)add\s+(column\s+)?\S+\s+[^;]*?not\s+null`)
	if reAddNotNull.MatchString(content) {
		if !regexp.MustCompile(`(?i)default\s+`).MatchString(content) {
			results = append(results, AuditResult{
				File:      filePath,
				CheckItem: "Column Constraints",
				Status:    StatusWarning,
				Details:   "Column added as NOT NULL without DEFAULT value.",
			})
		}
	}

	// Rule 4: CREATE TABLE without PRIMARY KEY or FK index (Warning)
	reCreateTable := regexp.MustCompile(`(?i)create\s+table\s+\S+\s*\(`)
	if reCreateTable.MatchString(content) {
		if !regexp.MustCompile(`(?i)primary\s+key`).MatchString(content) {
			results = append(results, AuditResult{
				File:      filePath,
				CheckItem: "Primary Key Check",
				Status:    StatusWarning,
				Details:   "CREATE TABLE without defining a Primary Key.",
			})
		}
	}

	// Default Green result if no errors/warnings found
	if len(results) == 0 {
		results = append(results, AuditResult{
			File:      filePath,
			CheckItem: "Destructive Commands",
			Status:    StatusGreen,
			Details:   "No destructive or inefficient operations detected.",
		})
	}

	return results, nil
}
