// main.go
package main

import (
	"os"

	"github.com/MeKo-Christian/go-sqlfmt/cmd" // import your cmd package
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
