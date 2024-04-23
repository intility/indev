/*
Copyright © 2024 Callum Powell <callum.powell@intility.no>
*/
package main

import (
	"github.com/intility/minctl/cmd"
)

//go:generate ./scripts/completions.sh

func main() {
	cmd.Execute()
}
