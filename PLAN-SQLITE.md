# SQLite Support – Implementation Plan

## Phase 1: Foundation (Language & Skeleton)

- [x] Add `SQLite` to `Language` in `config.go`.
- [x] Update `getFormatter()` in `format.go` to return a SQLite formatter.
- [x] Add `sqlite_formatter.go` with:
  - [x] `SQLiteFormatter` struct
  - [x] `NewSQLiteFormatter(cfg *Config)`
  - [x] `Format()` delegating to tokenizer + core formatter

- [x] Add `sqlite_formatter_test.go` with smoke tests.

## Phase 2: Tokenizer Config (baseline)

- [ ] `NewSQLiteTokenizerConfig()` seeded from Standard SQL:
  - [ ] **Comments**: `-- …` (line) and `/* … */` (block). No `#`. ([sqlite.org][2], [www2.sqlite.org][3])
  - [ ] **Identifiers**: accept `"double-quoted"` (standard), plus **backticks** and **\[brackets]** for compatibility. Keep as identifiers, not strings. ([O'Reilly Media][4], [Stack Overflow][5])
  - [ ] **Strings & blobs**: `'single-quoted'` strings; **blob** literals `X'ABCD'`.
  - [ ] **Numbers**: normal decimal/float; (no native `0b`).

- [ ] Tests: simple `SELECT` with `"double"`, `` `backtick` ``, `[bracket]`, comments.

## Phase 3: Placeholders (rich)

- [x] Support **all** SQLite bind parameter forms: `?`, `?NNN`, `:name`, `@name`, `$name` (1-based). Ensure tokenizer doesn't confuse them with `$` variables from other dialects. ([sqlite.org][6])
- [x] Update `params.go` for name/number resolution and 1-based indexing (leftmost is index 1). ([sqlite.org][7])
- [x] Tests: each style; mixtures; ensure placeholders inside strings/comments are ignored.

## Phase 4: Operators & Specials

- [x] **Concatenation**: `||` (treat as operator with normal spacing).
- [x] **JSON1** convenience ops: `->`, `->>` (SQLite ≥ 3.38); format like MySQL/PG analogs. ([sqlite.org][8], [SQLite Tutorial][9])
- [x] **NULL handling**: prefer `IS`, `IS NOT`; optionally recognize `IS [NOT] DISTINCT FROM` (SQLite ≥ 3.39). ([Stack Overflow][10])
- [x] **Pattern matching**: `LIKE`, `GLOB`; keep as reserved words (no special regex operator—`REGEXP` exists only via user function; treat as identifier/operator token without extra rules).
- [x] Tests: JSON chains, `a || b`, `x IS y`, `x IS NOT DISTINCT FROM y`.

## Phase 5: Core Clauses

- [x] **LIMIT**: support both `LIMIT n OFFSET m` and `LIMIT m, n` (both valid in SQLite). Keep one house style in output. ([sqlite.org][11])
- [x] **UPSERT** (SQLite ≥ 3.24): `INSERT … ON CONFLICT(col[, …]) DO UPDATE SET …` and `DO NOTHING`; also `INSERT OR REPLACE`. Tokenize/indent like PG's `ON CONFLICT` but with SQLite grammar. ([sqlite.org][11])
- [x] **Rowid knobs**: recognize `WITHOUT ROWID` as table option.
- [x] Tests: both LIMIT styles; `INSERT OR REPLACE`; `INSERT … ON CONFLICT … DO UPDATE`.

## Phase 6: CTEs & Window Functions

- [x] `WITH`, `WITH RECURSIVE` as top-level words; indentation mirrors PG. ([sqlite.org][12])
- [x] Windowing (`OVER`, `PARTITION BY`, frames) supported (3.25+ / 3.28+ features); format like PG. ([sqlite.org][13])
- [x] Tests: recursive CTE; simple window with `ROWS/RANGE` frame.

## Phase 7: DDL Essentials

- [x] Accept SQLite-flavored DDL:
  - [x] `CREATE TABLE` with **generated columns** (`GENERATED ALWAYS AS (expr) [VIRTUAL|STORED]`) and `STRICT`.
  - [x] `CREATE INDEX` (no engine/methods), `IF NOT EXISTS`.
- [x] `PRAGMA` as a **top-level** statement (format minimal; don’t reflow RHS).

- [x] Tests: table with generated cols; `CREATE INDEX IF NOT EXISTS`; simple `PRAGMA` lines.
- [x] Keep column/type lists conservative—SQLite’s type system is permissive; don’t attempt normalization. ([sqlite.org][14])

## Phase 8: Triggers & Views (lightweight)

- [x] `CREATE TRIGGER … BEGIN … END` and `CREATE VIEW … AS …`:
  - [x] Indent trigger bodies with `BEGIN/END`; avoid deep re-tokenizing inside body.

- [x] Tests: BEFORE/AFTER trigger skeleton; `CREATE VIEW` with CTE.

## Phase 9: Snapshot & Integration Tests

- [x] Add SQLite to snapshot suite (go-snaps). Golden file should include:
  - [x] JSON ops (`->`, `->>`), placeholders in multiple styles,
  - [x] LIMIT both styles,
  - [x] UPSERT variants,
  - [x] CTE + window example,
  - [x] DDL with generated columns/STRICT, and a `PRAGMA`.

- [x] Just targets: `just test-snapshots sqlite`, `just update-snapshots sqlite`.

## Phase 10: Docs

- [x] README: add SQLite to supported dialects with examples and note required SQLite versions for JSON/UPSERT/windows. ([sqlite.org][8])
- [x] Dialect notes:
  - [x] Comments supported; identifier quoting (`"…"`, `` `…` ``, `[ … ]`).
  - [x] Placeholders (`?`, `?NNN`, `:name`, `@name`, `$name`).
  - [x] Known limits: treat `REGEXP` like a bare identifier/operator; no pragma/value validation.

## Phase 11: Final Polish & Edge Cases

- [x] Keep `PRAGMA` values as-is (don't uppercase within string/identifier contexts).
- [x] Ensure semicolons inside strings/comments don't split statements (already in tokenizer, double-check). ([sqlite.org][15])
- [x] Unicode identifiers preserved; do not coerce case inside quotes.
- [x] Very large queries & malformed input fuzz tests.

---

### Minimal “definition of done”

- Dialect selectable as `sqlite`.
- Tokenizer handles comments, identifier quoting styles, blobs, and all placeholder forms.
- Formatter understands LIMIT styles, CTEs/windows, UPSERT, JSON operators.
- Snapshot suite green; README updated.

[1]: https://github.com/maxrichie5/go-sqlfmt?utm_source=chatgpt.com "GitHub - maxrichie5/go-sqlfmt: An SQL formatter written in Go."
[2]: https://sqlite.org/lang_comment.html?utm_source=chatgpt.com "SQL Comment Syntax - SQLite"
[3]: https://www2.sqlite.org/syntax/comment-syntax.html?utm_source=chatgpt.com "SQLite Syntax: comment-syntax"
[4]: https://www.oreilly.com/library/view/using-sqlite/9781449394592/ch04s03.html?utm_source=chatgpt.com "General Syntax - Using SQLite [Book] - O'Reilly Media"
[5]: https://stackoverflow.com/questions/75229248/what-do-square-brackets-around-an-identifier-mean-in-sqlite?utm_source=chatgpt.com "What do square brackets around an identifier mean in SQLite?"
[6]: https://sqlite.org/c3ref/bind_blob.html?utm_source=chatgpt.com "Binding Values To Prepared Statements - SQLite"
[7]: https://sqlite.org/c3ref/bind_parameter_name.html?utm_source=chatgpt.com "Name Of A Host Parameter - SQLite"
[8]: https://sqlite.org/json1.html?utm_source=chatgpt.com "JSON Functions And Operators - SQLite"
[9]: https://www.sqlitetutorial.net/sqlite-json-functions/sqlite-json-operators/?utm_source=chatgpt.com "SQLite JSON Operators"
[10]: https://stackoverflow.com/questions/9658125/what-is-the-equivalent-of-the-null-safe-equality-operator-in-sqlite?utm_source=chatgpt.com "What is the equivalent of the null-safe equality operator <=> in SQLite?"
[11]: https://sqlite.org/lang.html?utm_source=chatgpt.com "Query Language Understood by SQLite"
[12]: https://sqlite.org/lang_with.html?utm_source=chatgpt.com "The WITH Clause - SQLite"
[13]: https://sqlite.org/windowfunctions.html?utm_source=chatgpt.com "Window Functions - SQLite"
[14]: https://sqlite.org/quirks.html?utm_source=chatgpt.com "Quirks, Caveats, and Gotchas In SQLite"
[15]: https://sqlite.org/search?i=0&q=quoting&utm_source=chatgpt.com "Search SQLite Documentation"
[16]: https://sqlite.org/lang_keywords.html?utm_source=chatgpt.com "SQLite Keywords"
