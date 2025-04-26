package main

import (
	"fmt"
	"os"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/goat"
	"github.com/spf13/cobra"
)

func trackCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "track",
		Short: "Insert instrumentation code for the project",
		Long: `The track command is used to analyze incremental code of the project and insert instrumentation.

Options:
   None

Examples:
	goat track`,
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
				log.Errorf("Failed to load config: %v", err)
				return err
			}
			// check if the project is already patched
			if _, err := os.Stat(cfg.GoatGeneratedFile()); err == nil {
				return fmt.Errorf("project is already patched, please run `goat clean` first")
			}

			executor := goat.NewTrackExecutor(cfg)
			err = executor.Run()
			if err != nil {
				return err
			}

			// Show success message and next steps
			log.Info("----------------------------------------------------------")
			log.Info("✅ Track completed successfully!")
			log.Info("You can:")
			log.Info("- Review the changes using git diff or your preferred diff tool")
			log.Info("- Build and test your application to verify instrumentation")
			log.Info("- If you manualy add or remove instrumentation, run 'goat patch' to update the instrumentation")
			log.Info("- To remove all instrumentation, run 'goat clean'")
			log.Info("----------------------------------------------------------")

			return nil
		},
	}
	return cmd
}
