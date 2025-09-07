package sqlfmt

import (
	"fmt"
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// Clean up obsolete snapshots
	dirty, err := snaps.Clean(m)
	if err != nil {
		fmt.Println("Error cleaning snaps:", err)
		os.Exit(1)
	}
	if dirty {
		fmt.Println("Some snapshots were outdated.")
		os.Exit(1)
	}

	os.Exit(v)
}

func TestSnapshotFormatting_StandardSQL(t *testing.T) {
	formatter := NewStandardSQLFormatter(NewDefaultConfig())

	t.Run("basic SELECT", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = true;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("INSERT statement", func(t *testing.T) {
		query := "INSERT INTO users (name, email) VALUES ('John', 'john@test.com');"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}

func TestSnapshotFormatting_PostgreSQL(t *testing.T) {
	formatter := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL))

	t.Run("basic SELECT", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = true;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("JSON operations", func(t *testing.T) {
		query := "SELECT data->>'name' as name FROM users WHERE data ? 'active';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}

func TestSnapshotFormatting_N1QL(t *testing.T) {
	formatter := NewN1QLFormatter(NewDefaultConfig().WithLang(N1QL))

	t.Run("basic N1QL SELECT", func(t *testing.T) {
		query := "SELECT name FROM `travel-sample` WHERE type = 'airline';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}

func TestSnapshotFormatting_DB2(t *testing.T) {
	formatter := NewDB2Formatter(NewDefaultConfig().WithLang(DB2))

	t.Run("DB2 basic query", func(t *testing.T) {
		query := "SELECT empno, lastname FROM employee WHERE workdept = 'A00';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}

func TestSnapshotFormatting_PLSQL(t *testing.T) {
	formatter := NewPLSQLFormatter(NewDefaultConfig().WithLang(PLSQL))

	t.Run("PL/SQL basic query", func(t *testing.T) {
		query := "SELECT employee_id, last_name FROM employees WHERE department_id = 10;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}

func TestSnapshotFormatting_MySQL(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("basic SELECT", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = true;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("JSON operations", func(t *testing.T) {
		query := "SELECT profile->'$.name', settings->>'$.theme' FROM users WHERE data->'$.active' = 'true';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("NULL-safe equality", func(t *testing.T) {
		query := "SELECT * FROM users u LEFT JOIN profiles p ON u.id <=> p.user_id;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("REGEXP operations", func(t *testing.T) {
		query := "SELECT name FROM users WHERE email REGEXP '^[a-z]+@' AND status NOT REGEXP '^(banned|suspended)$';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("LIMIT variations", func(t *testing.T) {
		query := "SELECT * FROM products ORDER BY price LIMIT 20, 10;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT INTO products (id, name, price) VALUES (1, 'Product A', 19.99) ON DUPLICATE KEY UPDATE name = VALUES(name), price = VALUES(price), updated_at = NOW();"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("CTE with window function", func(t *testing.T) {
		query := "WITH sales_data AS (SELECT product_id, amount, ROW_NUMBER() OVER (PARTITION BY product_id ORDER BY amount DESC) as rank FROM sales) SELECT * FROM sales_data WHERE rank <= 3;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("DDL with generated columns", func(t *testing.T) {
		query := "CREATE TABLE orders (id INT PRIMARY KEY, subtotal DECIMAL(10,2), tax_rate DECIMAL(3,4), total DECIMAL(10,2) GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED);"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("stored procedure", func(t *testing.T) {
		query := "CREATE PROCEDURE GetUserStats(IN user_id INT) BEGIN SELECT COUNT(*) as orders, SUM(total) as revenue FROM orders WHERE customer_id = user_id; END;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("MySQL comments and backticks", func(t *testing.T) {
		query := "SELECT /*! SQL_CALC_FOUND_ROWS */ `user_id`, # hash comment\n\"full_name\" FROM `user_table` -- standard comment\nWHERE `active` = TRUE;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}

func TestSnapshotFormatting_WithCustomConfig(t *testing.T) {
	t.Run("uppercase keywords", func(t *testing.T) {
		cfg := NewDefaultConfig().WithUppercase()
		formatter := NewStandardSQLFormatter(cfg)
		query := "select id, name from users;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}
