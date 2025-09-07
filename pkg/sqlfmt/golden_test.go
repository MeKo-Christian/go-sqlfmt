package sqlfmt

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoldenFiles_StandardSQL(t *testing.T) {
	testGoldenFiles(t, "standard_sql", NewStandardSQLFormatter(NewDefaultConfig()))
}

func TestGoldenFiles_PostgreSQL(t *testing.T) {
	testGoldenFiles(t, "postgresql", NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)))
}

func TestGoldenFiles_N1QL(t *testing.T) {
	testGoldenFiles(t, "n1ql", NewN1QLFormatter(NewDefaultConfig().WithLang(N1QL)))
}

func TestGoldenFiles_DB2(t *testing.T) {
	testGoldenFiles(t, "db2", NewDB2Formatter(NewDefaultConfig().WithLang(DB2)))
}

func TestGoldenFiles_PLSQL(t *testing.T) {
	testGoldenFiles(t, "plsql", NewPLSQLFormatter(NewDefaultConfig().WithLang(PLSQL)))
}

func TestGoldenFiles_MySQL(t *testing.T) {
	testGoldenFiles(t, "mysql", NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)))
}

func testGoldenFiles(t *testing.T, dialect string, formatter Formatter) {
	t.Helper()

	// Get the directory of the current file
	var pc uintptr
	var line int
	pc, filename, line, ok := runtime.Caller(0)
	require.True(t, ok)
	_ = pc
	_ = line
	testDir := filepath.Dir(filename)

	// Navigate up to the project root and then to testdata
	projectRoot := filepath.Dir(filepath.Dir(testDir))
	inputDir := filepath.Join(projectRoot, "testdata", "input", dialect)
	goldenDir := filepath.Join(projectRoot, "testdata", "golden", dialect)

	// Walk through all .sql files in the input directory
	err := filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-SQL files
		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Get relative path for test name
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// Create test name from file path
		testName := strings.TrimSuffix(relPath, ".sql")
		testName = strings.ReplaceAll(testName, "/", "_")
		testName = strings.ReplaceAll(testName, "\\", "_")

		t.Run(testName, func(t *testing.T) {
			// Read the input file (unformatted SQL)
			inputBytes, err := os.ReadFile(path)
			require.NoError(t, err, "Failed to read input file %s", path)

			// Read the expected golden output
			goldenPath := filepath.Join(goldenDir, relPath)
			expectedBytes, err := os.ReadFile(goldenPath)
			require.NoError(t, err, "Failed to read golden file %s", goldenPath)

			// Format the input SQL
			actual := formatter.Format(string(inputBytes))
			expected := strings.TrimSpace(string(expectedBytes))

			// Compare formatted result against golden file
			require.Equal(t, expected, actual,
				"Formatted SQL doesn't match golden file.\nInput: %s\nGolden: %s\nActual:\n%s\nExpected:\n%s",
				path, goldenPath, actual, expected)
		})

		return nil
	})

	require.NoError(t, err, "Failed to walk input directory %s", inputDir)
}
