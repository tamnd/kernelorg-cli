package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (a *App) latestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "latest",
		Short: "Show the latest stable kernel version",
		Long:  `Print the latest stable Linux kernel version string to stdout and exit.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ls, err := a.client.LatestStable(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			_, _ = fmt.Fprintln(os.Stdout, ls.Version)
			return nil
		},
	}
}
