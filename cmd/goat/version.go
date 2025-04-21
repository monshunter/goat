package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information
var (
	Version   = "1.0.0"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Goat",
		Long:  `All software has versions. This is Goat's.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Goat version %s\n", Version)
			fmt.Printf("Commit: %s\n", Commit)
			fmt.Printf("Build date: %s\n", BuildDate)
		},
	}

	return cmd
}
