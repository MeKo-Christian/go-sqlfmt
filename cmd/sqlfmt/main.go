package main

import (
	"os"

	"github.com/maxrichie5/go-sqlfmt/cmd/sqlfmt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
