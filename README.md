# Backend Guardrails

Backend Guardrails is a set of lightweight, zero-dependency tools designed to complement `mattpocock/skills` by adding database schema safety and high-performance benchmarking capabilities directly to your AI agent workflows.

These tools act as pre-commit/pre-triage guardrails to prevent AI agents from generating destructive SQL migrations or introducing performance regressions.

## Included Guardrails

### 🛠️ 1. [migration-guard](./migration-guard)
A database schema safety guardrail written in Go.
* **Prevents Destructive Changes**: Detects `DROP TABLE`, `DROP COLUMN`, and `RENAME COLUMN` operations before deployment.
* **Enforces Rollbacks**: Ensures matching and non-empty `.down.sql` rollback files are present.
* **Rule Compliance**: Validates column constraints (like missing defaults on `NOT NULL` columns) and schema structures.

### ⚡ 2. [perf-audit](./perf-audit)
A performance benchmarking and comparative auditing tool for backend systems, allowing developers to monitor and review performance metrics.

---

For setup and usage instructions, please refer to the respective subdirectories.
