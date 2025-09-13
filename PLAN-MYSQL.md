# MySQL Support – Implementation Plan

## Phase 1: Foundation (Language & Skeleton)

- [x] Add `MySQL` to the `Language` type in `config.go`
- [x] Update `getFormatter()` in `format.go` to return a MySQL formatter
- [x] Add `mysql_formatter.go` with:
  - [x] `MySQLFormatter` struct
  - [x] `NewMySQLFormatter(cfg *Config)` ctor
  - [x] `Format()` method piping through tokenizer + core formatter

- [x] Add empty `mysql_formatter_test.go` (smoke tests scaffold)

> Notes: mirror the shape you used for PostgreSQL to keep parity.

## Phase 2: MySQL Tokenizer Config (baseline)

- [x] `NewMySQLTokenizerConfig()`:
  - [x] Start from Standard SQL config
  - [x] **Comments**: support `-- …`, `# …`, `/* … */`, and **versioned** comments `/*! … */` (treat as comments but keep intact) ([dev.mysql.com][2], [documentation.help][3])
  - [x] **Identifiers**: enable backtick-quoted identifiers (`` `name` ``)
  - [x] **Strings**: single-quoted strings; (optional) accept double quotes as strings unless you plan to honor `ANSI_QUOTES` later
  - [x] **Numbers/Literals**: allow `0xFF` hex, `0b1010` bit; `TRUE`/`FALSE`

- [x] Tests:
  - [x] init + simple `SELECT` with backticks, `#` comment, and `/*! hint */`

## Phase 3: Placeholders

- [x] Configure placeholders for Go MySQL drivers: `?` positional only (no `$1`)
- [x] Ensure placeholder regex does **not** misinterpret `?` in other contexts
- [x] Tests:
  - [x] `WHERE id = ?`
  - [x] Multiple `?` ordering
  - [x] Edge: `?` inside strings/comments ignored

## Phase 4: Operators & Special Tokens

- [x] **JSON operators**: `->` and `->>` (JSON extract and unquote) with normal spacing rules ([dev.mysql.com][4], [Oracle Docs][5])
- [x] **NULL-safe equality**: `<=>` (treat like a comparison operator) ([dev.mysql.com][6])
- [x] **Regex**: `REGEXP` / `RLIKE` and `NOT REGEXP` as reserved infix operators (tokenize as multi-word unit for `NOT REGEXP`) ([dev.mysql.com][7])
- [x] **Bitwise**: `|`, `&`, `^`, `~`, `<<`, `>>`
- [x] **Concatenation**: do **not** special-case `||` as concat (default MySQL treats `||` as logical OR unless `PIPES_AS_CONCAT` mode)
- [x] Tests:
  - [x] JSON extraction chains `doc->'$.a'->>'$.b'`
  - [x] `<=>` in joins/filters
  - [x] `col REGEXP '^foo|bar$'`

## Phase 5: Core Clauses (LIMIT, LOCKING, IGNORE)

- [x] **LIMIT**: normalize both forms: `LIMIT n OFFSET m` and `LIMIT m, n` (keep whichever style you choose; just format consistently) ([dev.mysql.com][8], [mysqltutorial.org][9], [DataCamp][10])
- [x] **Row locking**: `FOR UPDATE`, `FOR SHARE` (and legacy `LOCK IN SHARE MODE`—format but don't warn)
- [x] **INSERT variations**: `INSERT IGNORE`, `REPLACE` (keywords list only here)
- [x] Tests:
  - [x] `ORDER BY … LIMIT 10 OFFSET 20`
  - [x] `ORDER BY … LIMIT 20, 10`
  - [x] `SELECT … FOR UPDATE`

## Phase 6: MySQL "Upsert"

- [x] Add reserved sequence for `ON DUPLICATE KEY UPDATE` and keep it on one line break boundary (similar to `RETURNING` in PG plan)
- [x] Indentation rule: break before the clause when lines exceed width; indent the assignments list
- [x] Tests:
  - [x] Basic `INSERT … ON DUPLICATE KEY UPDATE col = VALUES(col)`
  - [x] Multiple assignments over lines

## Phase 7: CTEs & Window Functions (8.0+)

- [x] Add `WITH`, `WITH RECURSIVE` (top-level), same indentation you built for PG
- [x] Windowing: `OVER`, `PARTITION BY`, frame units (`RANGE`, `ROWS`)—mostly reusable
- [x] Tests:
  - [x] Non-recursive and recursive CTE
  - [x] Simple window function + frame

## Phase 8: DDL Essentials

- [x] Indexes: `CREATE [UNIQUE|FULLTEXT|SPATIAL] INDEX … USING BTREE|HASH`
- [x] Table options on `CREATE/ALTER TABLE`: tolerate/keep `ALGORITHM=INSTANT|INPLACE`, `LOCK=NONE|SHARED|EXCLUSIVE` (don't reflow inside `()` options too aggressively)
- [x] Generated columns: `col type GENERATED ALWAYS AS (expr) [VIRTUAL|STORED]`
- [x] Tests:
  - [x] `CREATE INDEX … USING BTREE`
  - [x] `ALTER TABLE … ALGORITHM=INSTANT, LOCK=NONE`

## Phase 9: Stored Routines (Minimal)

- [x] Recognize `CREATE PROCEDURE/FUNCTION … BEGIN … END`
- [x] Keep **routine body indentation** lightweight: indent blocks on `BEGIN/END`, `IF/ELSE`, `LOOP`
- [x] **Do not** attempt to manage `DELIMITER` (treat lines starting with `DELIMITER` as pass-through; no formatting inside) — this avoids breaking copy/paste flows
- [x] Tests:
  - [x] Procedure with simple control flow
  - [x] Ensure body strings/comments aren't retokenized oddly

> If your tokenizer tends to split on `;`, treat `DELIMITER` lines as turning off statement splitting until the next `DELIMITER`. If that’s non-trivial, mark deep routine formatting as **PARTIAL** and keep body mostly opaque, like your PL/pgSQL strategy.

## Phase 10: Snapshot & Integration Tests

- [x] Add MySQL to the snapshot suite you set up (go-snaps)
- [x] Golden file of "realish" MySQL 8.0 queries:
  - [x] JSON usage with `->`/`->>`
  - [x] Upserts
  - [x] `LIMIT` both styles
  - [x] CTEs + windows
  - [x] A couple of DDLs with options

- [x] Justfile targets: `just test-snapshots mysql`, `just update-snapshots mysql`

## Phase 11: Documentation

- [x] README: add MySQL to supported dialects list with examples & CLI flags (match upstream style) ([GitHub][1])
- [x] Brief notes on:
  - [x] Comments supported (incl. `/*! … */`)
  - [x] Placeholders (`?`)
  - [x] Known limitations (no `DELIMITER` management; no `ANSI_QUOTES` mode)

## Phase 12: Final Polish & Edge Cases

- [x] Backtick identifiers with unicode (including emoji) remain intact
- [x] Ensure `/*! … */` preserved verbatim (not reflowed) ([dev.mysql.com][2])
- [x] `NOT REGEXP` kept as a single logical unit for spacing
- [x] Keep `CONCAT()` as function (don't invent `||`)
- [x] Added hex/bit literal forms in tokenizer (`X'ABCD'`, `B'1010'`) in addition to `0x…/0b…`

---

## Minimal “definition of done”

- Formatters: Standard SQL unchanged; new MySQL dialect selectable
- MySQL basics: backticks, comments (incl. versioned), `?` placeholders
- Operators: JSON `->`, `->>`, null-safe `<=>`, `REGEXP`/`NOT REGEXP`
- Clauses: both `LIMIT` styles; `ON DUPLICATE KEY UPDATE`
- CTE/window: works like PG (only keywords differ)
- DDL: basic `CREATE INDEX` & `ALTER TABLE` options preserved
- Docs + snapshot tests green

[1]: https://github.com/maxrichie5/go-sqlfmt/blob/main/README.md?utm_source=chatgpt.com "go-sqlfmt/README.md at main · maxrichie5/go-sqlfmt · GitHub"
[2]: https://dev.mysql.com/doc/refman/8.4/en/comments.html?utm_source=chatgpt.com "MySQL :: MySQL 8.4 Reference Manual :: 11.7 Comments"
[3]: https://documentation.help/MySQL-5.0/ch09s04.html?utm_source=chatgpt.com "9.4. Comment Syntax - MySQL 5.0 Documentation"
[4]: https://dev.mysql.com/doc/refman/8.0/en/json.html?utm_source=chatgpt.com "MySQL :: MySQL 8.0 Reference Manual :: 13.5 The JSON Data Type"
[5]: https://docs.oracle.com/cd/E17952_01/mysql-8.0-en/json-functions.html?utm_source=chatgpt.com "14.17 JSON Functions - Oracle"
[6]: https://dev.mysql.com/doc/refman/8.4/en/comparison-operators.html?utm_source=chatgpt.com "14.4.2 Comparison Functions and Operators - MySQL"
[7]: https://dev.mysql.com/doc/refman/8.4/en/regexp.html?utm_source=chatgpt.com "MySQL :: MySQL 8.4 Reference Manual :: 14.8.2 Regular Expressions"
[8]: https://dev.mysql.com/doc/refman/8.0/en/limit-optimization.html?utm_source=chatgpt.com "10.2.1.19 LIMIT Query Optimization - MySQL"
[9]: https://www.mysqltutorial.org/mysql-basics/mysql-limit/?utm_source=chatgpt.com "MySQL LIMIT"
[10]: https://www.datacamp.com/doc/mysql/mysql-limit?utm_source=chatgpt.com "MySQL LIMIT Clause: Usage & Examples - DataCamp"
