# MySQL Support – Implementation Plan

## Phase 1: Foundation (Language & Skeleton)

- [ ] Add `MySQL` to the `Language` type in `config.go`
- [ ] Update `getFormatter()` in `format.go` to return a MySQL formatter
- [ ] Add `mysql_formatter.go` with:
  - [ ] `MySQLFormatter` struct
  - [ ] `NewMySQLFormatter(cfg *Config)` ctor
  - [ ] `Format()` method piping through tokenizer + core formatter

- [ ] Add empty `mysql_formatter_test.go` (smoke tests scaffold)

> Notes: mirror the shape you used for PostgreSQL to keep parity.

## Phase 2: MySQL Tokenizer Config (baseline)

- [ ] `NewMySQLTokenizerConfig()`:
  - [ ] Start from Standard SQL config
  - [ ] **Comments**: support `-- …`, `# …`, `/* … */`, and **versioned** comments `/*! … */` (treat as comments but keep intact) ([dev.mysql.com][2], [documentation.help][3])
  - [ ] **Identifiers**: enable backtick-quoted identifiers (`` `name` ``)
  - [ ] **Strings**: single-quoted strings; (optional) accept double quotes as strings unless you plan to honor `ANSI_QUOTES` later
  - [ ] **Numbers/Literals**: allow `0xFF` hex, `0b1010` bit; `TRUE`/`FALSE`

- [ ] Tests:
  - [ ] init + simple `SELECT` with backticks, `#` comment, and `/*! hint */`

## Phase 3: Placeholders

- [ ] Configure placeholders for Go MySQL drivers: `?` positional only (no `$1`)
- [ ] Ensure placeholder regex does **not** misinterpret `?` in other contexts
- [ ] Tests:
  - [ ] `WHERE id = ?`
  - [ ] Multiple `?` ordering
  - [ ] Edge: `?` inside strings/comments ignored

## Phase 4: Operators & Special Tokens

- [ ] **JSON operators**: `->` and `->>` (JSON extract and unquote) with normal spacing rules ([dev.mysql.com][4], [Oracle Docs][5])
- [ ] **NULL-safe equality**: `<=>` (treat like a comparison operator) ([dev.mysql.com][6])
- [ ] **Regex**: `REGEXP` / `RLIKE` and `NOT REGEXP` as reserved infix operators (tokenize as multi-word unit for `NOT REGEXP`) ([dev.mysql.com][7])
- [ ] **Bitwise**: `|`, `&`, `^`, `~`, `<<`, `>>`
- [ ] **Concatenation**: do **not** special-case `||` as concat (default MySQL treats `||` as logical OR unless `PIPES_AS_CONCAT` mode)
- [ ] Tests:
  - [ ] JSON extraction chains `doc->'$.a'->>'$.b'`
  - [ ] `<=>` in joins/filters
  - [ ] `col REGEXP '^foo|bar$'`

## Phase 5: Core Clauses (LIMIT, LOCKING, IGNORE)

- [ ] **LIMIT**: normalize both forms: `LIMIT n OFFSET m` and `LIMIT m, n` (keep whichever style you choose; just format consistently) ([dev.mysql.com][8], [mysqltutorial.org][9], [DataCamp][10])
- [ ] **Row locking**: `FOR UPDATE`, `FOR SHARE` (and legacy `LOCK IN SHARE MODE`—format but don’t warn)
- [ ] **INSERT variations**: `INSERT IGNORE`, `REPLACE` (keywords list only here)
- [ ] Tests:
  - [ ] `ORDER BY … LIMIT 10 OFFSET 20`
  - [ ] `ORDER BY … LIMIT 20, 10`
  - [ ] `SELECT … FOR UPDATE`

## Phase 6: MySQL “Upsert”

- [ ] Add reserved sequence for `ON DUPLICATE KEY UPDATE` and keep it on one line break boundary (similar to `RETURNING` in PG plan)
- [ ] Indentation rule: break before the clause when lines exceed width; indent the assignments list
- [ ] Tests:
  - [ ] Basic `INSERT … ON DUPLICATE KEY UPDATE col = VALUES(col)`
  - [ ] Multiple assignments over lines

## Phase 7: CTEs & Window Functions (8.0+)

- [ ] Add `WITH`, `WITH RECURSIVE` (top-level), same indentation you built for PG
- [ ] Windowing: `OVER`, `PARTITION BY`, frame units (`RANGE`, `ROWS`)—mostly reusable
- [ ] Tests:
  - [ ] Non-recursive and recursive CTE
  - [ ] Simple window function + frame

## Phase 8: DDL Essentials

- [ ] Indexes: `CREATE [UNIQUE|FULLTEXT|SPATIAL] INDEX … USING BTREE|HASH`
- [ ] Table options on `CREATE/ALTER TABLE`: tolerate/keep `ALGORITHM=INSTANT|INPLACE`, `LOCK=NONE|SHARED|EXCLUSIVE` (don’t reflow inside `()` options too aggressively)
- [ ] Generated columns: `col type GENERATED ALWAYS AS (expr) [VIRTUAL|STORED]`
- [ ] Tests:
  - [ ] `CREATE INDEX … USING BTREE`
  - [ ] `ALTER TABLE … ALGORITHM=INSTANT, LOCK=NONE`

## Phase 9: Stored Routines (Minimal)

- [ ] Recognize `CREATE PROCEDURE/FUNCTION … BEGIN … END`
- [ ] Keep **routine body indentation** lightweight: indent blocks on `BEGIN/END`, `IF/ELSE`, `LOOP`
- [ ] **Do not** attempt to manage `DELIMITER` (treat lines starting with `DELIMITER` as pass-through; no formatting inside) — this avoids breaking copy/paste flows
- [ ] Tests:
  - [ ] Procedure with simple control flow
  - [ ] Ensure body strings/comments aren’t retokenized oddly

> If your tokenizer tends to split on `;`, treat `DELIMITER` lines as turning off statement splitting until the next `DELIMITER`. If that’s non-trivial, mark deep routine formatting as **PARTIAL** and keep body mostly opaque, like your PL/pgSQL strategy.

## Phase 10: Snapshot & Integration Tests

- [ ] Add MySQL to the snapshot suite you set up (go-snaps)
- [ ] Golden file of “realish” MySQL 8.0 queries:
  - [ ] JSON usage with `->`/`->>`
  - [ ] Upserts
  - [ ] `LIMIT` both styles
  - [ ] CTEs + windows
  - [ ] A couple of DDLs with options

- [ ] Justfile targets: `just test-snapshots mysql`, `just update-snapshots mysql`

## Phase 11: Documentation

- [ ] README: add MySQL to supported dialects list with examples & CLI flags (match upstream style) ([GitHub][1])
- [ ] Brief notes on:
  - [ ] Comments supported (incl. `/*! … */`)
  - [ ] Placeholders (`?`)
  - [ ] Known limitations (no `DELIMITER` management; no `ANSI_QUOTES` mode)

## Phase 12: Final Polish & Edge Cases

- [ ] Backtick identifiers with unicode (including emoji) remain intact
- [ ] Ensure `/*! … */` preserved verbatim (not reflowed) ([dev.mysql.com][2])
- [ ] `NOT REGEXP` kept as a single logical unit for spacing
- [ ] Keep `CONCAT()` as function (don’t invent `||`)
- [ ] Optional: tolerate hex/bit literal forms in tokenizer (`X'ABCD'`, `B'1010'`) in addition to `0x…/0b…`

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
