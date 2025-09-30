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
		result := formatter.Format(basicSelectQuery)
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
		result := formatter.Format(basicSelectQuery)
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
		result := formatter.Format(basicSelectQuery)
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

func TestSnapshotFormatting_SQLite(t *testing.T) {
	formatter := NewSQLiteFormatter(NewDefaultConfig().WithLang(SQLite))

	t.Run("basic SELECT", func(t *testing.T) {
		result := formatter.Format(basicSelectQuery)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("JSON operations", func(t *testing.T) {
		query := "SELECT data->>'name' as name, profile->'settings' FROM users WHERE data->'active' = 'true';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("all placeholder styles", func(t *testing.T) {
		query := "SELECT * FROM users WHERE id = ? AND name = :name AND email = @email AND status = $status AND created_at > ?2;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("LIMIT variations", func(t *testing.T) {
		query1 := "SELECT * FROM products ORDER BY price LIMIT 10 OFFSET 20;"
		result1 := formatter.Format(query1)
		snaps.MatchSnapshot(t, result1)

		query2 := "SELECT * FROM products ORDER BY price LIMIT 20, 10;"
		result2 := formatter.Format(query2)
		snaps.MatchSnapshot(t, result2)
	})

	t.Run("UPSERT with ON CONFLICT", func(t *testing.T) {
		query := "INSERT INTO products (id, name, price) VALUES (1, 'Product A', 19.99) ON CONFLICT(id) DO UPDATE SET name = excluded.name, price = excluded.price;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("INSERT OR REPLACE", func(t *testing.T) {
		query := "INSERT OR REPLACE INTO cache (key, value, expires_at) VALUES ('user:123', '{\"name\":\"John\"}', datetime('now', '+1 day'));"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("CTE with RECURSIVE", func(t *testing.T) {
		query := "WITH RECURSIVE factorial(n, fact) AS (SELECT 1, 1 UNION ALL SELECT n+1, (n+1)*fact FROM factorial WHERE n < 10) SELECT * FROM factorial;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("window functions", func(t *testing.T) {
		query := "SELECT employee_id, department, salary, ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) as rank FROM employees;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("CTE with window functions", func(t *testing.T) {
		query := "WITH sales_data AS (SELECT product_id, amount, ROW_NUMBER() OVER (PARTITION BY product_id ORDER BY amount DESC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) as rank FROM sales) SELECT * FROM sales_data WHERE rank <= 3;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("DDL with generated columns", func(t *testing.T) {
		query := "CREATE TABLE orders (id INTEGER PRIMARY KEY, subtotal REAL, tax_rate REAL DEFAULT 0.08, total REAL GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED) STRICT;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("CREATE INDEX IF NOT EXISTS", func(t *testing.T) {
		query := "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE active = 1;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("PRAGMA statements", func(t *testing.T) {
		query := "PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("CREATE TRIGGER", func(t *testing.T) {
		query := "CREATE TRIGGER update_modified_time AFTER UPDATE ON users FOR EACH ROW BEGIN UPDATE users SET modified_at = datetime('now') WHERE id = NEW.id; END;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("CREATE VIEW with CTE", func(t *testing.T) {
		query := "CREATE VIEW active_users_summary AS WITH user_stats AS (SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id) SELECT u.name, us.order_count FROM users u JOIN user_stats us ON u.id = us.user_id WHERE u.active = 1;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("identifier quoting styles", func(t *testing.T) {
		query := "SELECT \"double_quoted\", `backtick_quoted`, [bracket_quoted] FROM \"table name\" WHERE `field name` = 'value';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("blob literals and concatenation", func(t *testing.T) {
		query := "SELECT name || ' - ' || description as full_name, X'DEADBEEF' as binary_data FROM products WHERE data = X'48656C6C6F';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("NULL handling", func(t *testing.T) {
		query := "SELECT * FROM users WHERE email IS NOT NULL AND status IS DISTINCT FROM 'deleted' AND name IS NOT DISTINCT FROM 'admin';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("GLOB and LIKE patterns", func(t *testing.T) {
		query := "SELECT * FROM files WHERE name GLOB '*.txt' OR path LIKE '%/temp/%' AND name NOT GLOB 'temp*';"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("WITHOUT ROWID table", func(t *testing.T) {
		query := "CREATE TABLE lookup (key TEXT PRIMARY KEY, value TEXT) WITHOUT ROWID;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("complex integrated example", func(t *testing.T) {
		query := `
WITH RECURSIVE category_tree AS (
  SELECT id, name, parent_id, 0 as level
  FROM categories WHERE parent_id IS NULL
  UNION ALL
  SELECT c.id, c.name, c.parent_id, ct.level + 1
  FROM categories c
  JOIN category_tree ct ON c.parent_id = ct.id
),
sales_summary AS (
  SELECT 
    p.category_id,
    COUNT(*) as order_count,
    SUM(oi.quantity * oi.price) as total_revenue,
    ROW_NUMBER() OVER (ORDER BY SUM(oi.quantity * oi.price) DESC) as revenue_rank
  FROM order_items oi
  JOIN products p ON oi.product_id = p.id
  WHERE oi.created_at >= date('now', '-1 month')
  GROUP BY p.category_id
)
SELECT 
  ct.name || ' (Level ' || ct.level || ')' as category_name,
  ss.order_count,
  ss.total_revenue,
  ss.revenue_rank
FROM category_tree ct
JOIN sales_summary ss ON ct.id = ss.category_id
WHERE ss.revenue_rank <= 10
ORDER BY ss.revenue_rank;`
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
