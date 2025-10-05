# go-sqlfmt Implementation Plan

This consolidated plan covers all remaining work for MySQL, PostgreSQL, and SQLite dialects, plus strategic enhancements for production readiness.

## Phase 1: Advanced SQL Features

### 1.1 Enhanced Stored Procedure Support

- [ ] PostgreSQL PL/pgSQL improvements:
  - [ ] Better block indentation (BEGIN/END, IF/ELSE, LOOP)
  - [ ] Exception handling formatting
  - [ ] Variable declaration formatting
- [ ] MySQL stored procedure improvements:
  - [ ] Better DELIMITER handling
  - [ ] Control flow structure formatting
  - [ ] Cursor handling
- [ ] SQLite trigger improvements:
  - [ ] Multi-statement trigger bodies
  - [ ] BEFORE/AFTER/INSTEAD OF formatting

### 1.2 Complete DDL Support

- [ ] Full CREATE TABLE support:
  - [ ] All constraint types
  - [ ] Table inheritance (PostgreSQL)
  - [ ] Partitioning clauses
  - [ ] Storage parameters
- [ ] Index creation options:
  - [ ] Partial indexes
  - [ ] Expression indexes
  - [ ] Index storage parameters
- [ ] ALTER TABLE statements:
  - [ ] All modification types
  - [ ] Multi-column changes
  - [ ] Constraint modifications
