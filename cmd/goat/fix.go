package main

import (
	"fmt"
	"os"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/goat"
	"github.com/spf13/cobra"
)

func fixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Process manual instrumentation markers in the project",
		Long: `The fix command is used to process manual instrumentation markers in the project.

It primarily handles:
  - // +goat:delete markers - removes code segments marked for deletion
  - // +goat:insert markers - inserts code at marked positions

Examples:
  goat fix`,
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
			executor := goat.NewFixExecutor(cfg)
			if err := executor.Run(); err != nil {
				return err
			}

			// 显示成功提示和后续建议
			log.Info("----------------------------------------------------------")
			log.Info("✅ Fix completed successfully!")
			log.Info("Manual markers have been processed (// +goat:delete, // +goat:insert)")
			log.Info("You can:")
			log.Info("- Review the changes using git diff or your preferred diff tool")
			log.Info("- Build and test your application to verify the changes")
			log.Info("- Add more manual markers and run 'goat fix' again if needed")
			log.Info("- If you want to remove all instrumentation, run 'goat clean'")
			log.Info("----------------------------------------------------------")

			return nil
		},
	}

	return cmd
}
