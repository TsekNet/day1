package cmd

import (
	"fmt"
	"runtime"

	"github.com/TsekNet/day1/internal/version"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, _ []string) {
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "day1 %s\n", version.Version)
			fmt.Fprintf(w, "  commit:  %s\n", version.Commit)
			fmt.Fprintf(w, "  built:   %s\n", version.Date)
			fmt.Fprintf(w, "  go:      %s\n", runtime.Version())
			fmt.Fprintf(w, "  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
