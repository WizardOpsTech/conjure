package repo

import (
	"github.com/spf13/cobra"
)

var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository management commands",
	Long: `Manage template and bundle repositories.

The repo command provides tools for creating and managing template and bundle
repositories, including generating index files for remote hosting.`,
}

func init() {
	RepoCmd.AddCommand(indexCmd)
}
