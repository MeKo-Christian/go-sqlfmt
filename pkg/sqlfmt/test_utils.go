package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCTEs runs common CTE (Common Table Expression) tests for any SQL formatter.
func TestCTEs(t *testing.T, formatter Formatter, lang Language) {
	testCTESimple(t, formatter)
	testCTERecursive(t, formatter)
	testCTEMultiple(t, formatter)
	testCTEComplexRecursive(t, formatter, lang)
}

func testCTESimple(t *testing.T, formatter Formatter) {
	t.Helper()
	t.Run("formats simple WITH query", func(t *testing.T) {
		query := "WITH user_count AS (SELECT COUNT(*) as total FROM users) SELECT total FROM user_count;"
		exp := Dedent(`
            WITH
              user_count AS (
                SELECT
                  COUNT(*) as total
                FROM
                  users
              )
            SELECT
              total
            FROM
              user_count;
        `)
		result := formatter.Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func testCTERecursive(t *testing.T, formatter Formatter) {
	t.Helper()
	t.Run("formats WITH RECURSIVE query", func(t *testing.T) {
		query := Dedent(`
            WITH RECURSIVE employee_tree AS (
                SELECT id, name, manager_id, 1 as level FROM employees WHERE manager_id IS NULL
                UNION ALL
                SELECT e.id, e.name, e.manager_id, et.level + 1
                FROM employees e JOIN employee_tree et ON e.manager_id = et.id
            ) SELECT * FROM employee_tree ORDER BY level, name;
        `)
		exp := Dedent(`
            WITH RECURSIVE
              employee_tree AS (
                SELECT
                  id,
                  name,
                  manager_id,
                  1 as level
                FROM
                  employees
                WHERE
                  manager_id IS NULL
                UNION ALL
                SELECT
                  e.id,
                  e.name,
                  e.manager_id,
                  et.level + 1
                FROM
                  employees e
                  JOIN employee_tree et ON e.manager_id = et.id
              )
            SELECT
              *
            FROM
              employee_tree
            ORDER BY
              level,
              name;
        `)
		result := formatter.Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func testCTEMultiple(t *testing.T, formatter Formatter) {
	t.Helper()
	t.Run("formats multiple CTEs", func(t *testing.T) {
		query := Dedent(`
            WITH active_users AS (SELECT * FROM users WHERE active = true),
                 recent_orders AS (SELECT * FROM orders WHERE created_at > '2023-01-01')
            SELECT u.name, o.total FROM active_users u JOIN recent_orders o ON u.id = o.user_id;
        `)
		exp := Dedent(`
            WITH
              active_users AS (
                SELECT
                  *
                FROM
                  users
                WHERE
                  active = true
              ),
              recent_orders AS (
                SELECT
                  *
                FROM
                  orders
                WHERE
                  created_at > '2023-01-01'
              )
            SELECT
              u.name,
              o.total
            FROM
              active_users u
              JOIN recent_orders o ON u.id = o.user_id;
        `)
		result := formatter.Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func testCTEComplexRecursive(t *testing.T, formatter Formatter, lang Language) {
	t.Helper()
	switch lang {
	case MySQL:
		testCTEComplexRecursiveMySQL(t, formatter)
	case PostgreSQL:
		testCTEComplexRecursivePostgreSQL(t, formatter)
	}
}

func testCTEComplexRecursiveMySQL(t *testing.T, formatter Formatter) {
	t.Helper()
	t.Run("formats complex recursive CTE with unions", func(t *testing.T) {
		query := Dedent(`
            WITH RECURSIVE category_hierarchy AS (
                SELECT id, name, parent_id, 0 as depth, JSON_ARRAY(id) as path
                FROM categories WHERE parent_id IS NULL
                UNION
                SELECT c.id, c.name, c.parent_id, ch.depth + 1, JSON_ARRAY_APPEND(ch.path, '$', c.id)
                FROM categories c JOIN category_hierarchy ch ON c.parent_id = ch.id
                WHERE NOT JSON_CONTAINS(ch.path, CAST(c.id AS JSON))
            ) SELECT id, name, depth, path FROM category_hierarchy;
        `)
		exp := Dedent(`
            WITH RECURSIVE
              category_hierarchy AS (
                SELECT
                  id,
                  name,
                  parent_id,
                  0 as depth,
                  JSON_ARRAY(id) as path
                FROM
                  categories
                WHERE
                  parent_id IS NULL
                UNION
                SELECT
                  c.id,
                  c.name,
                  c.parent_id,
                  ch.depth + 1,
                  JSON_ARRAY_APPEND(ch.path, '$', c.id)
                FROM
                  categories c
                  JOIN category_hierarchy ch ON c.parent_id = ch.id
                WHERE
                  NOT JSON_CONTAINS(ch.path, CAST(c.id AS JSON))
              )
            SELECT
              id,
              name,
              depth,
              path
            FROM
              category_hierarchy;
        `)
		result := formatter.Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func testCTEComplexRecursivePostgreSQL(t *testing.T, formatter Formatter) {
	t.Helper()
	t.Run("formats complex recursive CTE with unions", func(t *testing.T) {
		query := Dedent(`
            WITH RECURSIVE category_hierarchy AS (
                SELECT id, name, parent_id, 0 as depth, ARRAY[id] as path
                FROM categories WHERE parent_id IS NULL
                UNION
                SELECT c.id, c.name, c.parent_id, ch.depth + 1, ch.path || c.id
                FROM categories c JOIN category_hierarchy ch ON c.parent_id = ch.id
                WHERE NOT (c.id = ANY(ch.path))
            ) SELECT id, name, depth, array_to_string(path, ' > ') as breadcrumb FROM category_hierarchy;
        `)
		exp := Dedent(`
            WITH RECURSIVE
              category_hierarchy AS (
                SELECT
                  id,
                  name,
                  parent_id,
                  0 as depth,
                  ARRAY [id] as path
                FROM
                  categories
                WHERE
                  parent_id IS NULL
                UNION
                SELECT
                  c.id,
                  c.name,
                  c.parent_id,
                  ch.depth + 1,
                  ch.path || c.id
                FROM
                  categories c
                  JOIN category_hierarchy ch ON c.parent_id = ch.id
                WHERE
                  NOT (c.id = ANY(ch.path))
              )
            SELECT
              id,
              name,
              depth,
              array_to_string(path, ' > ') as breadcrumb
            FROM
              category_hierarchy;
        `)
		result := formatter.Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}
