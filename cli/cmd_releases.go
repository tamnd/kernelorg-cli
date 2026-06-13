package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) releasesCmd() *cobra.Command {
	var moniker string

	cmd := &cobra.Command{
		Use:   "releases",
		Short: "List kernel releases, optionally filtered by moniker",
		Long: `List all Linux kernel releases from kernel.org.

Use --moniker to filter to a specific release type:
  stable     the current stable release
  longterm   long-term support releases
  mainline   the current development kernel
  linux-next integration tree for the next merge window
  snapshot   periodic snapshots`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if moniker != "" {
				a.progressf("fetching %s releases...", moniker)
			} else {
				a.progressf("fetching all kernel releases...")
			}
			releases, err := a.client.Releases(cmd.Context(), moniker)
			if err != nil {
				return mapFetchErr(err)
			}
			releases = applyLimit(releases, a.limit)
			return a.renderOrEmpty(releases, len(releases))
		},
	}

	cmd.Flags().StringVar(&moniker, "moniker", "", "filter by moniker: stable|longterm|mainline|linux-next|snapshot")
	return cmd
}
