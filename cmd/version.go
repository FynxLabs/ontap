package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// versionCmd represents the version command
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Display version information for the OnTap CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("OnTap CLI v%s\n", Version)
			fmt.Printf("Build time: %s\n", BuildTime)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
