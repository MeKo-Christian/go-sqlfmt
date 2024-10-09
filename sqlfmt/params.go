package sqlfmt

import "strconv"

// params handles placeholder replacement with given parameters
type params struct {
	params Params
	index  int
}

// newParams creates a new params object
func newParams(p Params) *params {
	return &params{
		params: p,
		index:  0,
	}
}

func (p *params) emptyParams() bool {
	return len(p.params.MapParams) == 0 && len(p.params.ListParams) == 0
}

// get returns the param value that matches the given placeholder with param key.
// If a key is given, it first checks the MapParams for the value,
// and if it is not there, it will try to turn the key into an int which will be
// used as the index for the ListParams. If it is still not found, it returns
// the defaultValue. If the key is empty, it assumes you are using ListParams.
func (p *params) get(key string, defaultValue string) string {
	if p.emptyParams() {
		return defaultValue
	}

	if key != "" {
		if param, exists := p.params.MapParams[key]; exists {
			return param
		}

		if idx, err := strconv.Atoi(key); err == nil {
			if idx < len(p.params.ListParams) {
				return p.params.ListParams[idx]
			}
		}

		return defaultValue
	}

	if p.index >= len(p.params.ListParams) {
		return defaultValue
	}

	param := p.params.ListParams[p.index]
	p.index++
	return param
}
