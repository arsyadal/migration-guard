# Backend Guardrails

A collection of developer guardrail utilities for backend systems.

## Utilities

### 1. [migration-guard](./migration-guard)
A database schema safety guardrail written in Go. It scans SQL migration files, detects dangerous or inefficient SQL anti-patterns, checks for required primary keys/indexes, and verifies the presence of rollback (`.down.sql`) scripts.

### 2. [perf-audit](./perf-audit)
A performance benchmarking and comparative auditing tool for backend systems, allowing developers to monitor and review performance metrics.

---

For usage instructions of each tool, please refer to their respective directories.
