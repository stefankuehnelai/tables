// Command tables provides a CLI for Nextcloud Tables.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/stefankuehnelai/tables/internal/cli"
	"github.com/stefankuehnelai/tables/pkg/tables"
)

var (
	commandFactory = cli.NewRootCommand
	exitProcess    = os.Exit
)

func main() {
	exitProcess(run())
}

func run() int {
	command, err := commandFactory()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return exitCode(err)
	}
	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintln(command.ErrOrStderr(), err)
		return exitCode(err)
	}
	return tables.CodeOK
}

func exitCode(err error) int {
	var tablesError tables.TablesError
	if errors.As(err, &tablesError) {
		code := tablesError.Code()
		if code > 0 && code < 256 {
			return code
		}
	}
	return tables.CodeUnknown
}
