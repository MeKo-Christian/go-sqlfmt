package main

import (
	"fmt"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
)

func main() {
	cfg := sqlfmt.NewDefaultConfig().
		WithAlignAssignments(true)

	query := "UPDATE users SET name = 'John', email = 'john@example.com', age = 30 WHERE id = 1;"

	result := sqlfmt.Format(query, cfg)
	fmt.Println("Query:")
	fmt.Println(query)
	fmt.Println()
	fmt.Println("Formatted:")
	fmt.Println(result)

	// Test without alignment
	cfg2 := sqlfmt.NewDefaultConfig().
		WithAlignAssignments(false)

	result2 := sqlfmt.Format(query, cfg2)
	fmt.Println()
	fmt.Println("Without alignment:")
	fmt.Println(result2)
}
