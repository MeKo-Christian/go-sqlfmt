package main

import (
	"fmt"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
)

func main() {
	cfg := sqlfmt.NewDefaultConfig()
	cfg.AlignColumnNames = true

	result := sqlfmt.Format("SELECT id, name, email FROM users;", cfg)
	fmt.Println("With alignment:")
	fmt.Println(result)

	cfg.AlignColumnNames = false
	result = sqlfmt.Format("SELECT id, name, email FROM users;", cfg)
	fmt.Println("Without alignment:")
	fmt.Println(result)
}
