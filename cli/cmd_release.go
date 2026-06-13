package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) releaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release <version>",
		Short: "Show details for a specific kernel release",
		Long:  `Show full metadata for a single Linux kernel release identified by version string.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			a.progressf("fetching release %s...", version)
			r, err := a.client.GetRelease(cmd.Context(), version)
			if err != nil {
				return mapFetchErr(err)
			}
			if err := a.render(r); err != nil {
				return codeError(exitError, fmt.Errorf("render: %w", err))
			}
			return nil
		},
	}
}
