# 📊 perf-audit

> **Minimalist Performance & Allocation Guardrail for Go AI Agents.**

`perf-audit` is a zero-dependency, ultra-lightweight CLI tool written in **Go**, designed specifically to function as an **Agent Skill (`skills.sh`)**.

It acts as an automated guardrail to prevent AI Coding Agents (such as Claude Code, Cursor, or Windsurf) from writing inefficient, heap-allocated Go code or introducing performance regressions into your critical hot paths.

---

## 💡 Why This Matters

AI Coding Agents are excellent at writing functional code rapidly, but they are often **blind to low-level performance**. Agents frequently write non-idiomatic Go code, scatter unnecessary pointers, or trigger excessive heap allocations that silently degrade backend performance.

`perf-audit` forces the AI to prove that its code is not just **working**, but **fast and memory-efficient** before it is allowed to commit the changes.

---

## ✨ Key Features

- 🚀 **Zero External Dependencies:** Built purely on the Go standard library (`os/exec`, `regexp`, `encoding/json`). Fast compilation, near-zero overhead.
- 📉 **Automated Baseline Tracking:** Automatically captures stable performance metrics and persists them locally into a `.perf-baseline.json` file.
- 🛡️ **10% Regression Threshold:** Strictly exits with status code `1` if memory usage (`B/op`) or allocation count (`allocs/op`) increases by more than 10% compared to the baseline.
- 🤖 **Agent Native:** Outputs structured Markdown tables that coding agents can instantly parse, interpret, and act upon.

---

## 🛠️ Installation

### Build the binary

```bash
git clone https://github.com/arsyadal/perf-audit.git
cd perf-audit
go build -o perf-audit .
sudo mv perf-audit /usr/local/bin/   # optional: install globally
```

Or run without building:

```bash
go run /path/to/perf-audit/*.go
```

### Plug into your AI coding tool

Most tools auto-read `AGENTS.md` — copy it to your project root for zero-setup support.

```bash
# copy AGENTS.md into your Go project
cp /path/to/perf-audit/AGENTS.md ./AGENTS.md
```

| Tool | Setup |
| :--- | :--- |
| **Claude Code** | Copy `AGENTS.md` to project root. Or add `hooks/hooks.json` entries to your Claude Code settings for always-on injection. |
| **Codex** | Copy `AGENTS.md` — auto-loaded from repo root via VS Code Codex extension (`~/.codex/AGENTS.md` for global). |
| **Cursor** | Copy `.cursor/rules/perf-audit.md` to your project's `.cursor/rules/`. |
| **Windsurf** | Copy `.windsurf/rules/perf-audit.md` to your project's `.windsurf/rules/`. |
| **Kiro** | Copy `.kiro/steering/perf-audit.md` to `~/.kiro/steering/` (global) or `.kiro/steering/` in your project. |
| **GitHub Copilot Chat** | Copy `.github/copilot-instructions.md` to your project's `.github/`. |
| **Pi agent** | `SKILL.md` at repo root is Pi-compatible. Pi reads it automatically when you run from this directory. |
| **Amp (Sourcegraph)** | Copy `AGENTS.md` — auto-read from working dir and parents up to `$HOME`. |
| **Jules (Google)** | Copy `AGENTS.md` to repo root — auto-read. |
| **Aider** | Copy `AGENTS.md` to project root — auto-read as context. |
| **Zed** | Copy `AGENTS.md` to project root. |

> **Any tool that reads `AGENTS.md`** gets perf-audit context with zero config. Copy once, works everywhere.

---

## 🚀 Usage

Run from the root of any Go project that has `Benchmark*` functions in `*_test.go` files:

```bash
# Using compiled binary
/path/to/perf-audit

# Or go run
go run /path/to/perf-audit/*.go
```

---

## 📋 Example Output

**First run — no baseline exists:**

```
### 📊 Performance Audit Report

No baseline found. Saved current metrics as new baseline.

| Benchmark Name | Speed (ns/op) | Memory (B/op) | Allocs (allocs/op) |
| :--- | ---: | ---: | ---: |
| `BenchmarkProcessData` | 94 ns | 536 B | 2 |
```

**Subsequent run — regression detected:**

```
### 📊 Performance Audit Report

| Benchmark Name | Metric | Baseline | Current | Delta | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `BenchmarkProcessData` | Speed (`ns/op`) | 80 ns | 84 ns | +5.0% | ✅ Green |
|  | Memory (`B/op`) | 300 B | 536 B | +78.7% | ❌ Regression |
|  | Allocs (`allocs/op`) | 1 | 2 | +100.0% | ❌ Regression |

⚠️ **Audit Result:** Performance regression detected in memory metrics. Optimization is strictly required before committing.
```

**Clean run:**

```
✅ **Audit Result:** All metrics within threshold. No regression detected.
```

---

## 🔢 Exit Codes

| Code | Meaning |
| :--- | :--- |
| `0` | Clean — all metrics within threshold |
| `1` | Regression — `B/op` or `allocs/op` increased >10% |
| `2` | Tool error — no benchmarks found, build failed, parse error |

---

## ⚙️ How It Works

1. Runs `go test -bench=. -benchmem ./...` in current working directory
2. Parses `ns/op`, `B/op`, `allocs/op` from output via regex
3. **First run:** saves `.perf-baseline.json`, exits `0`
4. **Subsequent runs:** compares against baseline, exits `1` if regression detected

### Regression rule

Memory (`B/op`) or allocations (`allocs/op`) increasing **>10%** triggers exit code `1`.

Speed (`ns/op`) is reported but **not** used for regression gating — too noisy across environments.

---

## 📁 File Structure

```
perf-audit/
├── main.go               # Entry point, report printer, exit logic
├── benchmark_parser.go   # Regex parser for go test -benchmem output
├── comparator.go         # Baseline load/save, delta calc, threshold check
├── go.mod                # Zero external dependencies
├── SKILL.md              # Agent skill manifest (skills.sh ecosystem)
└── .gitignore            # Excludes .perf-baseline.json and binary
```

---

## 🗄️ Baseline File

Saved as `.perf-baseline.json` in the working directory. Gitignored by default.

```json
{
  "timestamp": "2026-07-17T15:00:00Z",
  "benchmarks": {
    "ProcessData": {
      "ns_op": 94,
      "bytes_op": 536,
      "allocs_op": 2
    }
  }
}
```

To reset baseline: delete `.perf-baseline.json` and rerun.

---

## 🤖 Agent Skill Contract

See [`SKILL.md`](./SKILL.md) for full agent invocation contract.

**TL;DR for agents:** If `perf-audit` exits `1`, locate the regressing benchmark, refactor to reduce heap allocations, and rerun. **Do not declare the task complete until exit code is `0`.**

---

## 📦 Requirements

- Go 1.21+
- Target project must have `*_test.go` files with `Benchmark*` functions using `-benchmem` compatible output
