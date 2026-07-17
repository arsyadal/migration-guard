---
name: perf-audit
description: Run local performance audits and memory allocation guardrails for Go codebases.
---

# Performance Audit Skill

Use this skill whenever you complete a task that alters hot paths, loops, database queries, or core algorithms.

## How to execute

Run the compiled binary from the root of the target Go project:

```bash
go run path/to/perf-audit/main.go
```

Or if compiled:

```bash
perf-audit
```

## Behavior

- **First run**: No `.perf-baseline.json` exists → saves current metrics as baseline, exits 0.
- **Subsequent runs**: Compares current benchmarks against baseline. Exits 0 if clean, exits 1 if regression detected.

## Regression rule

Memory (`B/op`) or allocations (`allocs/op`) increasing **>10%** triggers exit code 1.

## Agent contract

If the tool exits with status 1, read the output table, locate the regressing benchmark, and refactor to reduce heap allocations. **Do not declare the task complete until the audit passes (exit 0).**
