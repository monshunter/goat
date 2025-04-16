package main

import (
	"fmt"
	"log"
	"os"

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
			if _, err := os.Stat(configYaml); os.IsNotExist(err) {
				return fmt.Errorf("config file %s not found", configYaml)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configYaml)
			if err != nil {
				log.Printf("failed to load config: %v", err)
				return err
			}
			cleanExecutor := goat.NewCleanExecutor(cfg)
			return cleanExecutor.Run()
		},
	}

	return cmd
}
