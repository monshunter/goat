package main

import (
	"fmt"
	"os"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/goat"
	"github.com/spf13/cobra"
)

func patchCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Insert instrumentation code for the project",
		Long: `The patch command is used to analyze incremental code of the project and insert instrumentation.

Options:
   None

Examples:
	goat patch`,
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
			// check if the project is already patched
			if _, err := os.Stat(cfg.GoatGeneratedFile()); err == nil {
				return fmt.Errorf("project is already patched, please run `goat clean` first")
			}

			executor := goat.NewPatchExecutor(cfg)
			return executor.Run()
		},
	}
	return cmd
}
