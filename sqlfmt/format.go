package sqlfmt

import "fmt"

func Format(query string, cfg ...Config) string {
	if len(cfg) == 1 {
		switch cfg[0].Language {
		default:
			return NewStandardSQLFormatter(cfg[0]).Format(query)
		}
	}

	if len(cfg) > 1 {
		panic("cannot have more than one config")
	}

	return NewStandardSQLFormatter(NewDefaultConfig()).Format(query)
}

func PrettyPrint(query string, cfg ...Config) {
	// TODO: colors
	// TODO: cfg
	fmt.Println(NewStandardSQLFormatter(NewDefaultConfig()).Format(query))
}
