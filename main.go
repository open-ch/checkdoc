package main

import (
	"log/slog"
	"os"

	"github.com/open-ch/checkdoc/cmd"
)

func main() {
	if err := cmd.GetRootCommand().Execute(); err != nil {
		slog.Error("checkdoc failed", "err", err)
		os.Exit(1)
	}
}
