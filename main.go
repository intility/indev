/*
Copyright © 2024 Callum Powell <callum.powell@intility.no>
*/
package main

import (
	"os"

	"github.com/intility/icpctl/cmd"
)

//go:generate ./scripts/completions.sh

func main() {
	exitCode := run(os.Args)
	os.Exit(exitCode)
}

func run(args []string) int {
	return cmd.Execute(args)
}
