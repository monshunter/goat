package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean <project>",
		Short: "Clean up instrumentation code in project",
		Long: `The clean command is used to remove all instrumentation code from the project.

Arguments:
  <project> Project path

Examples:
  goat clean /path/to/project
  goat clean .`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]
			fmt.Printf("Cleaning project: %s\n", project)

			// TODO: Implement cleanup logic
			return nil
		},
	}

	return cmd
}
