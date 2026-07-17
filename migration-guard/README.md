# migration-guard

A database schema safety guardrail for AI agents. It scans SQL migration files, detects dangerous operations (e.g. `DROP` or `RENAME`), verifies rollback (`.down.sql`) files, and outputs a structured Markdown compliance report.

## Rules checked

* **Critical Hazards** (exits with code 1):
  * `DROP TABLE` or `DROP COLUMN`
  * `ALTER TABLE ... RENAME COLUMN`
  * Missing or empty corresponding `.down.sql` file
* **Warnings** (passes with warning):
  * `ADD COLUMN` with `NOT NULL` but missing a `DEFAULT` value
  * `CREATE TABLE` without defining a `PRIMARY KEY`

## Usage

Run the checker in your migrations directory:

```bash
go run . -dir ./db/migrations
```

### Options

* `-dir`: Path to the directory containing migration files (default: `./migrations`).

## License

MIT
