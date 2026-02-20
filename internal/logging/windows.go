//go:build windows

package logging

import (
	"fmt"
	"os"

	"github.com/google/deck"
	"github.com/google/deck/backends/eventlog"
)

func Init() {
	evt, err := eventlog.InitWithDefaultInstall("day1")
	if err != nil {
		evt, err = eventlog.Init("day1")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "eventlog init: %v (stderr only)\n", err)
		return
	}
	deck.Add(evt)
}
