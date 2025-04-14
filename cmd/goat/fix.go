package main

import (
	"fmt"
	"log"
	"os"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/goat"
	"github.com/spf13/cobra"
)

func fixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Fix instrumentation code in the project",
		Long: `The fix command is used to fix potentially problematic instrumentation code in the project.

Examples:
  goat fix`,
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
			return goat.RunFix(cfg)
		},
	}

	return cmd
}
