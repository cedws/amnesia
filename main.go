package main

import (
	"fmt"
	"os"

	"github.com/cedws/amnesia/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
