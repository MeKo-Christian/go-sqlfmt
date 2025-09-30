package utils

import "strconv"

// ParamsConfig represents the public params configuration.
type ParamsConfig struct {
	MapParams  map[string]string
	ListParams []string
	// UseSQLiteIndexing enables 1-based indexing for SQLite compatibility
	// When true, ?1 maps to ListParams[0], ?2 to ListParams[1], etc.
	// When false (default), uses 0-based indexing where ?0 maps to ListParams[0]
	UseSQLiteIndexing bool
}

// Params handles placeholder replacement with given parameters.
type Params struct {
	params *ParamsConfig
	index  int
}

// NewParams creates a new params object.
func NewParams(p *ParamsConfig) *Params {
	if p == nil {
		p = &ParamsConfig{}
	}
	if p.MapParams == nil {
		p.MapParams = make(map[string]string)
	}
	if p.ListParams == nil {
		p.ListParams = make([]string, 0)
	}
	return &Params{
		params: p,
		index:  0,
	}
}

func (p *Params) EmptyParams() bool {
	return len(p.params.MapParams) == 0 && len(p.params.ListParams) == 0
}

// Get returns the param value that matches the given placeholder with param key.
// If a key is given, it first checks the MapParams for the value,
// and if it is not there, it will try to turn the key into an int which will be
// used as the index for the ListParams. If it is still not found, it returns
// the defaultValue. If the key is empty, it assumes you are using ListParams.
// For SQLite compatibility, when UseSQLiteIndexing is true, numbered placeholders
// use 1-based indexing (?1 -> ListParams[0], ?2 -> ListParams[1], etc.)
func (p *Params) Get(key string, defaultValue string) string {
	if p.EmptyParams() {
		return defaultValue
	}

	if key != "" {
		return p.getByKey(key, defaultValue)
	}

	return p.getByIndex(defaultValue)
}

func (p *Params) getByKey(key string, defaultValue string) string {
	if param, exists := p.params.MapParams[key]; exists {
		return param
	}

	if idx, err := strconv.Atoi(key); err == nil {
		return p.getByNumericIndex(idx, defaultValue)
	}

	return defaultValue
}

func (p *Params) getByNumericIndex(idx int, defaultValue string) string {
	if p.params.UseSQLiteIndexing {
		// SQLite uses 1-based indexing: ?1 maps to ListParams[0]
		if idx >= 1 && idx-1 < len(p.params.ListParams) {
			return p.params.ListParams[idx-1]
		}
	} else {
		// Default 0-based indexing: ?0 maps to ListParams[0]
		if idx >= 0 && idx < len(p.params.ListParams) {
			return p.params.ListParams[idx]
		}
	}
	return defaultValue
}

func (p *Params) getByIndex(defaultValue string) string {
	if p.index >= len(p.params.ListParams) {
		return defaultValue
	}

	param := p.params.ListParams[p.index]
	p.index++
	return param
}
