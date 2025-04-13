package main

import (
	"log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <project> --stable <stableBranch> --publish <publishBranch>",
		Short: "Initialize a new project",
		Long: `The init command is used to initialize a new project.

Arguments:
  <project> Project path

Options:
  --stable <stableBranch> Stable branch
  --publish <publishBranch> Publish branch

Examples:
  goat init /path/to/project --stable master --publish "release-1.32"
  goat init .`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]
			log.Printf("Initializing project: %s\n", project)
			err := config.Init(project)
			if err != nil {
				log.Fatalf("failed to initialize project: %v", err)
				return err
			}
			return nil
		},
	}

	return cmd
}
