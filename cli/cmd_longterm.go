package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) longtermCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "longterm",
		Aliases: []string{"lts"},
		Short:   "List all active longterm (LTS) kernel releases",
		Long:    `List all active long-term support (LTS) Linux kernel releases.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching longterm releases...")
			releases, err := a.client.Releases(cmd.Context(), "longterm")
			if err != nil {
				return mapFetchErr(err)
			}
			releases = applyLimit(releases, a.limit)
			return a.renderOrEmpty(releases, len(releases))
		},
	}
}
