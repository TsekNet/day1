package main

import (
	"embed"
	"os"

	"github.com/TsekNet/day1/cmd"
	"github.com/TsekNet/day1/internal/logging"
	"github.com/google/deck"
	"github.com/google/deck/backends/logger"
)

//go:embed all:frontend
var assets embed.FS

//go:embed all:testdata/pages
var embeddedPages embed.FS

func main() {
	deck.Add(logger.Init(os.Stderr, 0))
	logging.Init()

	if len(os.Args) > 1 && os.Args[1] == "-generate-bindings" {
		cmd.RunBindings(assets)
		return
	}

	if err := cmd.Execute(assets, embeddedPages); err != nil {
		deck.Errorf("%v", err)
		os.Exit(1)
	}
}
