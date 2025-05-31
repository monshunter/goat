package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [flags]",
		Short: "Initialize a new project",
		Long: `The init command is used to initialize a new project in the current directory.

Options:
  --old <oldBranch>                     Old branch for comparison base (default: "main"), valid values: [commit hash, branch name, tag name, "", HEAD, INIT (for new repository)]
  --new <newBranch>                     New branch for comparison target (default: "HEAD"), valid values: [commit hash, branch name, tag name, "", HEAD]
  --app-name <appName>                  Application name (default: current directory name)
  --app-version <appVersion>            Application version (default: current commit short hash)
  --granularity <granularity>           Granularity (line, patch, scope, func) (default: "patch")
  --diff-precision <diffPrecision>      Diff precision (1~3) (default: 1)
  --threads <threads>                   Number of threads (default: 1)
  --race                                Enable race detection (default: false)
  --goat-package-name <packageName>     Goat package name (default: "goat")
  --goat-package-alias <packageAlias>   Goat package alias (default: "goat")
  --goat-package-path <packagePath>     Goat package path (default: "goat")
  --ignores <ignores>                   Comma-separated list of files/dirs to ignore
  --main-entries <entries>              Comma-separated list of relative paths to main packages from project root (e.g., 'cmd/server,cmd/client' or '*' for all)
  --printer-config-mode <mode>          Printer config mode, list of (none, useSpaces, tabIndent, sourcePos, rawFormat) (default: "useSpaces,tabIndent")
  --printer-config-tabwidth <tabwidth>  Printer config tabwidth (default: 8)
  --printer-config-indent <indent>      Printer config indent (default: 0)
  --data-type <dataType>                Data type (bool, count) (default: "bool")
  --skip-nested-modules                 Skip directories containing go.mod files (default: true)
  --force                               Force overwrite existing goat.yaml file

Examples:
  goat init --old master --new "release-1.32"
  goat init --app-name "my-app" --app-version "2.0.0" --granularity func
  goat init --threads 4 --race
  goat init --ignores ".git,.idea,node_modules"
  goat init --main-entries "cmd/app,cmd/worker"
  goat init --printer-config-mode "useSpaces,tabIndent" --printer-config-tabwidth 4 --printer-config-indent 2
  goat init --force                     Force overwrite existing goat.yaml file`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := "."
			if _, err := os.Stat(project); os.IsNotExist(err) {
				return fmt.Errorf("current directory not found")
			}
			log.Info("Project initializing ... ")
			project, err := filepath.Abs(project)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			// Check if config file already exists
			filename := config.ConfigYaml
			if !filepath.IsAbs(config.ConfigYaml) {
				filename = filepath.Join(project, config.ConfigYaml)
			}

			force, _ := cmd.Flags().GetBool("force")
			if _, err := os.Stat(filename); err == nil && !force {
				return fmt.Errorf("config file %s already exists. Use 'goat init --force' to overwrite or manually edit the file", filename)
			}

			// get all command line options
			oldBranch, _ := cmd.Flags().GetString("old")
			newBranch, _ := cmd.Flags().GetString("new")
			appName, _ := cmd.Flags().GetString("app-name")
			appVersion, _ := cmd.Flags().GetString("app-version")
			granularity, _ := cmd.Flags().GetString("granularity")
			diffPrecision, _ := cmd.Flags().GetInt("diff-precision")
			threads, _ := cmd.Flags().GetInt("threads")
			race, _ := cmd.Flags().GetBool("race")
			goatPackageName, _ := cmd.Flags().GetString("goat-package-name")
			goatPackageAlias, _ := cmd.Flags().GetString("goat-package-alias")
			goatPackagePath, _ := cmd.Flags().GetString("goat-package-path")
			ignoresStr, _ := cmd.Flags().GetString("ignores")
			mainEntriesStr, _ := cmd.Flags().GetString("main-entries")
			printerConfigModeStr, _ := cmd.Flags().GetString("printer-config-mode")
			printerConfigTabwidth, _ := cmd.Flags().GetInt("printer-config-tabwidth")
			printerConfigIndent, _ := cmd.Flags().GetInt("printer-config-indent")
			dataTypeStr, _ := cmd.Flags().GetString("data-type")
			skipNestedModules, _ := cmd.Flags().GetBool("skip-nested-modules")

			// process ignore file list
			var ignores []string
			ignoresStr = strings.TrimSpace(ignoresStr)
			if ignoresStr != "" {
				ignores = strings.Split(ignoresStr, ",")
			}
			// process main package list
			var mainEntries []string
			if mainEntriesStr != "" {
				mainEntries = strings.Split(mainEntriesStr, ",")
			} else {
				mainEntries = []string{"*"}
			}

			var printerConfigMode []config.PrinterConfigMode
			if printerConfigModeStr != "" {
				modes := strings.Split(printerConfigModeStr, ",")
				for _, mode := range modes {
					printerConfigMode = append(printerConfigMode, config.PrinterConfigMode(mode))
				}
			} else {
				printerConfigMode = []config.PrinterConfigMode{config.PrinterConfigModeUseSpaces, config.PrinterConfigModeTabIndent}
			}

			// create config
			cfg := &config.Config{
				AppName:               appName,
				AppVersion:            appVersion,
				OldBranch:             oldBranch,
				NewBranch:             newBranch,
				GoatPackageName:       goatPackageName,
				GoatPackageAlias:      goatPackageAlias,
				GoatPackagePath:       goatPackagePath,
				Threads:               threads,
				Race:                  race,
				Ignores:               ignores,
				MainEntries:           mainEntries,
				Granularity:           granularity,
				DiffPrecision:         diffPrecision,
				PrinterConfigMode:     printerConfigMode,
				PrinterConfigTabwidth: printerConfigTabwidth,
				PrinterConfigIndent:   printerConfigIndent,
				DataType:              dataTypeStr,
				SkipNestedModules:     skipNestedModules,
			}

			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("failed to validate config: %w", err)
			}

			err = config.InitWithConfig(filename, cfg)
			if err != nil {
				log.Fatalf("Failed to initialize project: %v", err)
				return err
			}

			log.Info("Project initialized successfully")
			log.Info("Suggested next steps:")
			log.Infof("- Open and Edit: '%s' if needed", filename)
			return nil
		},
	}

	// add command line options
	cmd.Flags().String("old", "main", "Old branch for comparison base, valid values: [commit hash, branch name, tag name, '', HEAD, INIT (for new repository)]")
	cmd.Flags().String("new", "HEAD", "New branch for comparison target, valid values: [commit hash, branch name, tag name, '', HEAD], newBranch must be the same as the current HEAD")
	cmd.Flags().String("app-name", "", "Application name")
	cmd.Flags().String("app-version", "", "Application version")
	cmd.Flags().String("granularity", "patch", "Granularity (line, patch, scope, func)")
	cmd.Flags().Int("diff-precision", 1, "Diff precision (1~3)")
	cmd.Flags().Int("threads", 1, "Number of threads")
	cmd.Flags().Bool("race", false, "Enable race detection")
	cmd.Flags().String("goat-package-name", "goat", "Goat package name")
	cmd.Flags().String("goat-package-alias", "goat", "Goat package alias")
	cmd.Flags().String("goat-package-path", "goat", "Goat package path")
	cmd.Flags().String("ignores", "", "Comma-separated list of files/dirs to ignore")
	cmd.Flags().String("main-entries", "", "Comma-separated list of relative paths to main packages from project root (e.g., 'cmd/server,cmd/client' or '*' for all)")
	cmd.Flags().String("printer-config-mode", "useSpaces,tabIndent", "Printer config mode, list of (none, useSpaces, tabIndent, sourcePos, rawFormat)")
	cmd.Flags().Int("printer-config-tabwidth", 8, "Printer config tabwidth")
	cmd.Flags().Int("printer-config-indent", 0, "Printer config indent")
	cmd.Flags().String("data-type", "bool", "Data type (bool, count)")
	cmd.Flags().Bool("skip-nested-modules", true, "Skip sub directories containing go.mod files")
	cmd.Flags().Bool("force", false, "Force overwrite existing goat.yaml file")

	return cmd
}
