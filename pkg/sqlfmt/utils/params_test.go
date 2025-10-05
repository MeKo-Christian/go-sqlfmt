package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewParams tests the NewParams constructor with various configurations.
func TestNewParams(t *testing.T) {
	tests := []struct {
		name     string
		config   *ParamsConfig
		validate func(*testing.T, *Params)
	}{
		{
			name:   "nil config creates empty params",
			config: nil,
			validate: func(t *testing.T, p *Params) {
				require.NotNil(t, p)
				require.NotNil(t, p.params)
				require.NotNil(t, p.params.MapParams)
				require.NotNil(t, p.params.ListParams)
				require.True(t, p.EmptyParams())
			},
		},
		{
			name: "config with MapParams only",
			config: &ParamsConfig{
				MapParams: map[string]string{"key": "value"},
			},
			validate: func(t *testing.T, p *Params) {
				require.False(t, p.EmptyParams())
				require.Equal(t, "value", p.params.MapParams["key"])
			},
		},
		{
			name: "config with ListParams only",
			config: &ParamsConfig{
				ListParams: []string{"first", "second", "third"},
			},
			validate: func(t *testing.T, p *Params) {
				require.False(t, p.EmptyParams())
				require.Len(t, p.params.ListParams, 3)
			},
		},
		{
			name: "config with both MapParams and ListParams",
			config: &ParamsConfig{
				MapParams:  map[string]string{"name": "John"},
				ListParams: []string{"value1", "value2"},
			},
			validate: func(t *testing.T, p *Params) {
				require.False(t, p.EmptyParams())
				require.Equal(t, "John", p.params.MapParams["name"])
				require.Len(t, p.params.ListParams, 2)
			},
		},
		{
			name: "config with SQLite indexing enabled",
			config: &ParamsConfig{
				ListParams:        []string{"a", "b", "c"},
				UseSQLiteIndexing: true,
			},
			validate: func(t *testing.T, p *Params) {
				require.True(t, p.params.UseSQLiteIndexing)
			},
		},
		{
			name:   "empty config initializes maps and slices",
			config: &ParamsConfig{},
			validate: func(t *testing.T, p *Params) {
				require.NotNil(t, p.params.MapParams)
				require.NotNil(t, p.params.ListParams)
				require.True(t, p.EmptyParams())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(tt.config)
			tt.validate(t, p)
		})
	}
}

// TestEmptyParams tests the EmptyParams method.
func TestEmptyParams(t *testing.T) {
	tests := []struct {
		name     string
		config   *ParamsConfig
		expected bool
	}{
		{
			name:     "nil config is empty",
			config:   nil,
			expected: true,
		},
		{
			name:     "empty MapParams and ListParams",
			config:   &ParamsConfig{},
			expected: true,
		},
		{
			name: "empty MapParams but non-empty ListParams",
			config: &ParamsConfig{
				ListParams: []string{"value"},
			},
			expected: false,
		},
		{
			name: "non-empty MapParams but empty ListParams",
			config: &ParamsConfig{
				MapParams: map[string]string{"key": "value"},
			},
			expected: false,
		},
		{
			name: "both MapParams and ListParams non-empty",
			config: &ParamsConfig{
				MapParams:  map[string]string{"key": "value"},
				ListParams: []string{"value"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(tt.config)
			result := p.EmptyParams()
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_EmptyParams tests Get method when params are empty.
func TestGet_EmptyParams(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "empty params with key returns default",
			key:          "someKey",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty params without key returns default",
			key:          "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty params with numeric key returns default",
			key:          "1",
			defaultValue: "fallback",
			expected:     "fallback",
		},
		{
			name:         "empty params with empty default",
			key:          "key",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(nil)
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_MapParams tests Get method with named parameters.
func TestGet_MapParams(t *testing.T) {
	tests := []struct {
		name         string
		mapParams    map[string]string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "get existing key",
			mapParams:    map[string]string{"username": "alice"},
			key:          "username",
			defaultValue: "default",
			expected:     "alice",
		},
		{
			name:         "get non-existing key returns default",
			mapParams:    map[string]string{"username": "alice"},
			key:          "password",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "get with empty key returns default",
			mapParams:    map[string]string{"username": "alice"},
			key:          "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "multiple keys - get first",
			mapParams:    map[string]string{"key1": "value1", "key2": "value2"},
			key:          "key1",
			defaultValue: "default",
			expected:     "value1",
		},
		{
			name:         "multiple keys - get second",
			mapParams:    map[string]string{"key1": "value1", "key2": "value2"},
			key:          "key2",
			defaultValue: "default",
			expected:     "value2",
		},
		{
			name:         "empty string value",
			mapParams:    map[string]string{"key": ""},
			key:          "key",
			defaultValue: "default",
			expected:     "",
		},
		{
			name:         "special characters in key",
			mapParams:    map[string]string{"@param": "value"},
			key:          "@param",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "colon-prefixed key",
			mapParams:    map[string]string{":id": "123"},
			key:          ":id",
			defaultValue: "default",
			expected:     "123",
		},
		{
			name:         "case-sensitive keys",
			mapParams:    map[string]string{"Key": "value1", "key": "value2"},
			key:          "key",
			defaultValue: "default",
			expected:     "value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(&ParamsConfig{
				MapParams: tt.mapParams,
			})
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_ListParams_ZeroBasedIndexing tests Get method with 0-based indexed parameters.
func TestGet_ListParams_ZeroBasedIndexing(t *testing.T) {
	tests := []struct {
		name         string
		listParams   []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "get first element with index 0",
			listParams:   []string{"first", "second", "third"},
			key:          "0",
			defaultValue: "default",
			expected:     "first",
		},
		{
			name:         "get second element with index 1",
			listParams:   []string{"first", "second", "third"},
			key:          "1",
			defaultValue: "default",
			expected:     "second",
		},
		{
			name:         "get third element with index 2",
			listParams:   []string{"first", "second", "third"},
			key:          "2",
			defaultValue: "default",
			expected:     "third",
		},
		{
			name:         "out of bounds index returns default",
			listParams:   []string{"first", "second"},
			key:          "5",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "negative index returns default",
			listParams:   []string{"first", "second"},
			key:          "-1",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "single element list",
			listParams:   []string{"only"},
			key:          "0",
			defaultValue: "default",
			expected:     "only",
		},
		{
			name:         "empty list returns default",
			listParams:   []string{},
			key:          "0",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(&ParamsConfig{
				ListParams:        tt.listParams,
				UseSQLiteIndexing: false, // 0-based indexing
			})
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_ListParams_SQLiteIndexing tests Get method with 1-based indexed parameters (SQLite mode).
func TestGet_ListParams_SQLiteIndexing(t *testing.T) {
	tests := []struct {
		name         string
		listParams   []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "get first element with index 1",
			listParams:   []string{"first", "second", "third"},
			key:          "1",
			defaultValue: "default",
			expected:     "first",
		},
		{
			name:         "get second element with index 2",
			listParams:   []string{"first", "second", "third"},
			key:          "2",
			defaultValue: "default",
			expected:     "second",
		},
		{
			name:         "get third element with index 3",
			listParams:   []string{"first", "second", "third"},
			key:          "3",
			defaultValue: "default",
			expected:     "third",
		},
		{
			name:         "index 0 returns default (SQLite starts at 1)",
			listParams:   []string{"first", "second"},
			key:          "0",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "out of bounds index returns default",
			listParams:   []string{"first", "second"},
			key:          "10",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "negative index returns default",
			listParams:   []string{"first", "second"},
			key:          "-1",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "single element list with index 1",
			listParams:   []string{"only"},
			key:          "1",
			defaultValue: "default",
			expected:     "only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(&ParamsConfig{
				ListParams:        tt.listParams,
				UseSQLiteIndexing: true, // 1-based indexing
			})
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_SequentialAccess tests Get method without key for sequential access.
func TestGet_SequentialAccess(t *testing.T) {
	t.Run("sequential access returns values in order", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams: []string{"first", "second", "third"},
		})

		require.Equal(t, "first", p.Get("", "default"))
		require.Equal(t, "second", p.Get("", "default"))
		require.Equal(t, "third", p.Get("", "default"))
	})

	t.Run("sequential access beyond list returns default", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams: []string{"first", "second"},
		})

		require.Equal(t, "first", p.Get("", "default"))
		require.Equal(t, "second", p.Get("", "default"))
		require.Equal(t, "default", p.Get("", "default"))
		require.Equal(t, "default", p.Get("", "default"))
	})

	t.Run("sequential access with empty list returns default", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams: []string{},
		})

		require.Equal(t, "default", p.Get("", "default"))
	})

	t.Run("sequential access increments index", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams: []string{"a", "b", "c"},
		})

		// Access sequentially
		p.Get("", "default")
		p.Get("", "default")

		// Third call should get third element
		require.Equal(t, "c", p.Get("", "default"))
	})
}

// TestGet_MixedAccess tests Get method with both named and indexed parameters.
func TestGet_MixedAccess(t *testing.T) {
	tests := []struct {
		name         string
		config       *ParamsConfig
		key          string
		defaultValue string
		expected     string
	}{
		{
			name: "named param takes precedence over indexed",
			config: &ParamsConfig{
				MapParams:  map[string]string{"0": "named_zero"},
				ListParams: []string{"indexed_zero"},
			},
			key:          "0",
			defaultValue: "default",
			expected:     "named_zero",
		},
		{
			name: "get named param when both exist",
			config: &ParamsConfig{
				MapParams:  map[string]string{"id": "123"},
				ListParams: []string{"first", "second"},
			},
			key:          "id",
			defaultValue: "default",
			expected:     "123",
		},
		{
			name: "get indexed param when name doesn't exist",
			config: &ParamsConfig{
				MapParams:  map[string]string{"name": "alice"},
				ListParams: []string{"first", "second"},
			},
			key:          "1",
			defaultValue: "default",
			expected:     "second",
		},
		{
			name: "sequential access ignores MapParams",
			config: &ParamsConfig{
				MapParams:  map[string]string{"key": "value"},
				ListParams: []string{"first"},
			},
			key:          "",
			defaultValue: "default",
			expected:     "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(tt.config)
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_EdgeCases tests edge cases for the Get method.
func TestGet_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		config       *ParamsConfig
		key          string
		defaultValue string
		expected     string
	}{
		{
			name: "whitespace in key",
			config: &ParamsConfig{
				MapParams: map[string]string{" key ": "value"},
			},
			key:          " key ",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name: "numeric string that looks like float",
			config: &ParamsConfig{
				ListParams: []string{"first", "second"},
			},
			key:          "1.5",
			defaultValue: "default",
			expected:     "default", // Should fail to parse as int
		},
		{
			name: "very large index",
			config: &ParamsConfig{
				ListParams: []string{"only"},
			},
			key:          "999999",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name: "unicode characters in map key",
			config: &ParamsConfig{
				MapParams: map[string]string{"日本語": "value"},
			},
			key:          "日本語",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name: "unicode characters in map value",
			config: &ParamsConfig{
				MapParams: map[string]string{"key": "こんにちは"},
			},
			key:          "key",
			defaultValue: "default",
			expected:     "こんにちは",
		},
		{
			name: "empty string in ListParams",
			config: &ParamsConfig{
				ListParams: []string{"", "second"},
			},
			key:          "0",
			defaultValue: "default",
			expected:     "",
		},
		{
			name: "null character in value",
			config: &ParamsConfig{
				MapParams: map[string]string{"key": "val\x00ue"},
			},
			key:          "key",
			defaultValue: "default",
			expected:     "val\x00ue",
		},
		{
			name: "key with leading zeros",
			config: &ParamsConfig{
				ListParams: []string{"a", "b", "c"},
			},
			key:          "01",
			defaultValue: "default",
			expected:     "b", // "01" parses to int 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(tt.config)
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestGet_DefaultValues tests various default value scenarios.
func TestGet_DefaultValues(t *testing.T) {
	tests := []struct {
		name         string
		config       *ParamsConfig
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "empty default value",
			config:       &ParamsConfig{},
			key:          "missing",
			defaultValue: "",
			expected:     "",
		},
		{
			name:         "whitespace default value",
			config:       &ParamsConfig{},
			key:          "missing",
			defaultValue: "   ",
			expected:     "   ",
		},
		{
			name:         "special characters in default",
			config:       &ParamsConfig{},
			key:          "missing",
			defaultValue: "!@#$%^&*()",
			expected:     "!@#$%^&*()",
		},
		{
			name:         "SQL injection attempt in default",
			config:       &ParamsConfig{},
			key:          "missing",
			defaultValue: "'; DROP TABLE users; --",
			expected:     "'; DROP TABLE users; --",
		},
		{
			name:         "long default value",
			config:       &ParamsConfig{},
			key:          "missing",
			defaultValue: "very_long_default_value_that_exceeds_normal_length",
			expected:     "very_long_default_value_that_exceeds_normal_length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams(tt.config)
			result := p.Get(tt.key, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestParamsThreadSafety tests that separate Params instances don't interfere.
func TestParamsThreadSafety(t *testing.T) {
	config := &ParamsConfig{
		ListParams: []string{"first", "second", "third"},
	}

	// Create two separate instances
	p1 := NewParams(config)
	p2 := NewParams(config)

	// Sequential access on p1 shouldn't affect p2
	require.Equal(t, "first", p1.Get("", "default"))
	require.Equal(t, "second", p1.Get("", "default"))

	// p2 should still get first element
	require.Equal(t, "first", p2.Get("", "default"))

	// Continue with p1
	require.Equal(t, "third", p1.Get("", "default"))

	// p2 should get second element
	require.Equal(t, "second", p2.Get("", "default"))
}

// TestGet_RealWorldSQLScenarios tests real-world SQL placeholder scenarios.
func TestGet_RealWorldSQLScenarios(t *testing.T) {
	t.Run("PostgreSQL positional parameters", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams:        []string{"John", "john@example.com", "2024-01-01"},
			UseSQLiteIndexing: true, // PostgreSQL uses $1, $2, $3 (1-based)
		})

		// $1, $2, $3 style placeholders
		require.Equal(t, "John", p.Get("1", "?"))
		require.Equal(t, "john@example.com", p.Get("2", "?"))
		require.Equal(t, "2024-01-01", p.Get("3", "?"))
	})

	t.Run("MySQL/SQLite positional parameters", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams: []string{"value1", "value2", "value3"},
		})

		// ? style placeholders (sequential access)
		require.Equal(t, "value1", p.Get("", "?"))
		require.Equal(t, "value2", p.Get("", "?"))
		require.Equal(t, "value3", p.Get("", "?"))
	})

	t.Run("SQLite numbered placeholders (1-based)", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			ListParams:        []string{"Alice", "Bob", "Charlie"},
			UseSQLiteIndexing: true,
		})

		// ?1, ?2, ?3 style placeholders
		require.Equal(t, "Alice", p.Get("1", "?"))
		require.Equal(t, "Bob", p.Get("2", "?"))
		require.Equal(t, "Charlie", p.Get("3", "?"))
	})

	t.Run("Named parameters", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			MapParams: map[string]string{
				"username": "alice",
				"email":    "alice@example.com",
				"role":     "admin",
			},
		})

		// :name or @name style placeholders
		require.Equal(t, "alice", p.Get("username", "?"))
		require.Equal(t, "alice@example.com", p.Get("email", "?"))
		require.Equal(t, "admin", p.Get("role", "?"))
	})

	t.Run("Mixed named and positional", func(t *testing.T) {
		p := NewParams(&ParamsConfig{
			MapParams: map[string]string{
				"table_name": "users",
				"limit":      "100",
			},
			ListParams: []string{"active", "verified"},
		})

		// Get named params
		require.Equal(t, "users", p.Get("table_name", "?"))
		require.Equal(t, "100", p.Get("limit", "?"))

		// Get positional params
		require.Equal(t, "active", p.Get("0", "?"))
		require.Equal(t, "verified", p.Get("1", "?"))
	})
}
