package main

import (
	"context"
	"fmt"
	"os"

	"github.com/katurdays/unconf/internal/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(cli.ExitCodeForError(err))
	}
}
