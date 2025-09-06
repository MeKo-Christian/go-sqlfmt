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

func TestSnapshotFormatting_WithCustomConfig(t *testing.T) {
	t.Run("uppercase keywords", func(t *testing.T) {
		cfg := NewDefaultConfig().WithUppercase()
		formatter := NewStandardSQLFormatter(cfg)
		query := "select id, name from users;"
		result := formatter.Format(query)
		snaps.MatchSnapshot(t, result)
	})
}
