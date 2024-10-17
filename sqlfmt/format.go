package sqlfmt

import "fmt"

type Formatter interface {
	Format(string) string
}

// Format formats the SQL query according to an optional config.
func Format(query string, cfg ...Config) string {
	return getFormatter(false, cfg...).Format(query)
}

// PrettyFormat formats the SQL query the same as Format but with coloring added.
func PrettyFormat(query string, cfg ...Config) string {
	return getFormatter(true, cfg...).Format(query)
}

// PrettyPrint calls PrettyFormat and prints the formatted query.
func PrettyPrint(query string, cfg ...Config) {
	fmt.Println(PrettyFormat(query, cfg...))
}

func getFormatter(forceWithColor bool, cfg ...Config) Formatter {
	c := NewDefaultConfig()

	if len(cfg) == 1 {
		c = cfg[0]
	}

	if len(cfg) > 1 {
		panic("cannot have more than one config")
	}

	if forceWithColor {
		c.ColorConfig = NewDefaultColorConfig()
	}

	switch c.Language {
	default:
		return NewStandardSQLFormatter(c)
	}
}
