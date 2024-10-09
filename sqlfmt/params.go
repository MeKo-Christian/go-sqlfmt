package sqlfmt

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
// If the param is missing, it returns the defaultValue.
// If the key is empty, it assumes you are using ListParams.
func (p *params) get(key string, defaultValue string) string {
	if p.emptyParams() {
		return defaultValue
	}

	if key != "" {
		if param, exists := p.params.MapParams[key]; exists {
			return param
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
