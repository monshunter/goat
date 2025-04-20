package main

import (
	"fmt"
	"os"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/goat"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up instrumentation code in project",
		Long: `The clean command is used to remove all instrumentation code from the project.

Examples:
  goat clean`,
		Args: cobra.ExactArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(config.ConfigYaml); os.IsNotExist(err) {
				return fmt.Errorf("config file %s not found", config.ConfigYaml)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(config.ConfigYaml)
			if err != nil {
				log.Errorf("failed to load config: %v", err)
				return err
			}
			cleanExecutor := goat.NewCleanExecutor(cfg)
			return cleanExecutor.Run()
		},
	}

	return cmd
}
