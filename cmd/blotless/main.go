package main

import (
	"io"
	"os"

	"github.com/blotless/cli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Stdin))
}

func run(args []string, stdout, stderr io.Writer, stdin io.Reader) int {
	return cli.Execute(args, cli.Options{
		Version: cli.Version,
		Stdout:  stdout,
		Stderr:  stderr,
		Stdin:   stdin,
	})
}
