/*
Copyright © 2024 Callum Powell <callum.powell@intility.no>
*/
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/intility/icpctl/cmd"
)

//go:generate ./scripts/completions.sh

func main() {
	exitCode := run(os.Args)
	os.Exit(exitCode)
}

func run(args []string) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		if <-sigChan; true {
			cancel()
		}
	}()

	return cmd.Execute(ctx, args)
}
