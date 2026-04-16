package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/katurdays/unconf/internal/servercli"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := servercli.Execute(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
