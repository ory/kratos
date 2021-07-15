package main

import (
	"fmt"
	"os"

	"github.com/ory/kratos/cmd"

	"github.com/ory/x/clidoc"
)

func main() {
	if err := clidoc.Generate(cmd.NewRootCmd(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
	fmt.Println("All files have been generated and updated.")
}
