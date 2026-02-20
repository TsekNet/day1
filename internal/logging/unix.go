//go:build !windows

package logging

import (
	"fmt"
	"os"

	"github.com/google/deck"
	"github.com/google/deck/backends/syslog"
)

func Init() {
	sl, err := syslog.Init("day1", syslog.LOG_USER)
	if err != nil {
		fmt.Fprintf(os.Stderr, "syslog init: %v (stderr only)\n", err)
		return
	}
	deck.Add(sl)
}
