package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/monshunter/goat/pkg/config"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <project> [flags]",
		Short: "Initialize a new project",
		Long: `The init command is used to initialize a new project.

Arguments:
  <project> Project path

Options:
  --stable <stableBranch>               Stable branch (default: "main")
  --publish <publishBranch>             Publish branch (default: ".")
  --app-name <appName>                  Application name (default: "example-app")
  --app-version <appVersion>            Application version (default: "1.0.0")
  --granularity <granularity>           Granularity (line, block, func) (default: "block")
  --diff-precision <diffPrecision>      Diff precision (1-2) (default: 1)
  --threads <threads>                   Number of threads (default: 1)
  --race                                Enable race detection (default: false)
  --clone-branch                        Clone branch (default: false)
  --goat-package-name <packageName>     Goat package name (default: "goat")
  --goat-package-alias <packageAlias>   Goat package alias (default: "goat")
  --goat-package-path <packagePath>     Goat package path (default: "goat")
  --track-strategy <strategy>           Main package track strategy (project, package) (default: "project")
  --ignores <ignores>                   Comma-separated list of files/dirs to ignore
  --main-entries <entries>              Comma-separated list of main entries to track (default: "*")

Examples:
  goat init /path/to/project --stable master --publish "release-1.32"
  goat init . --app-name "my-app" --app-version "2.0.0" --granularity func
  goat init . --threads 4 --race --clone-branch
  goat init . --ignores ".git,.idea,node_modules"
  goat init . --main-entries "cmd/app,cmd/worker"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]
			if _, err := os.Stat(project); os.IsNotExist(err) {
				return fmt.Errorf("project %s not found", project)
			}
			log.Printf("Initializing project: %s\n", project)
			project, err := filepath.Abs(project)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}
			// get all command line options
			stableBranch, _ := cmd.Flags().GetString("stable")
			publishBranch, _ := cmd.Flags().GetString("publish")
			appName, _ := cmd.Flags().GetString("app-name")
			appVersion, _ := cmd.Flags().GetString("app-version")
			granularity, _ := cmd.Flags().GetString("granularity")
			diffPrecision, _ := cmd.Flags().GetInt("diff-precision")
			threads, _ := cmd.Flags().GetInt("threads")
			race, _ := cmd.Flags().GetBool("race")
			cloneBranch, _ := cmd.Flags().GetBool("clone-branch")
			goatPackageName, _ := cmd.Flags().GetString("goat-package-name")
			goatPackageAlias, _ := cmd.Flags().GetString("goat-package-alias")
			goatPackagePath, _ := cmd.Flags().GetString("goat-package-path")
			trackStrategy, _ := cmd.Flags().GetString("track-strategy")
			ignoresStr, _ := cmd.Flags().GetString("ignores")
			mainEntriesStr, _ := cmd.Flags().GetString("main-entries")

			// process ignore file list
			var ignores []string
			if ignoresStr != "" {
				ignores = strings.Split(ignoresStr, ",")
			} else {
				// use default ignore file list
				ignores = []string{".git", ".gitignore", ".DS_Store", ".idea", ".vscode", ".venv"}
			}

			// process main package list
			var mainEntries []string
			if mainEntriesStr != "" {
				mainEntries = strings.Split(mainEntriesStr, ",")
			} else {
				mainEntries = []string{"*"}
			}

			// create config
			cfg := &config.Config{
				AppName:          appName,
				AppVersion:       appVersion,
				ProjectRoot:      project,
				StableBranch:     stableBranch,
				PublishBranch:    publishBranch,
				GoatPackageName:  goatPackageName,
				GoatPackageAlias: goatPackageAlias,
				GoatPackagePath:  goatPackagePath,
				Threads:          threads,
				Race:             race,
				CloneBranch:      cloneBranch,
				TrackStrategy:    trackStrategy,
				Ignores:          ignores,
				MainEntries:      mainEntries,
				Granularity:      granularity,
				DiffPrecision:    diffPrecision,
			}

			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("failed to validate config: %w", err)
			}

			filename := configYaml
			if !filepath.IsAbs(configYaml) {
				filename = filepath.Join(project, configYaml)
			}
			// initialize config
			err = config.InitWithConfig(filename, cfg)
			if err != nil {
				log.Fatalf("failed to initialize project: %v", err)
				return err
			}

			log.Printf("Project initialized successfully")
			return nil
		},
	}

	// add command line options
	cmd.Flags().String("stable", "main", "Stable branch")
	cmd.Flags().String("publish", "HEAD", "Publish branch")
	cmd.Flags().String("app-name", "example-app", "Application name")
	cmd.Flags().String("app-version", "1.0.0", "Application version")
	cmd.Flags().String("granularity", "block", "Granularity (line, block, func)")
	cmd.Flags().Int("diff-precision", 1, "Diff precision (1-4)")
	cmd.Flags().Int("threads", 1, "Number of threads")
	cmd.Flags().Bool("race", false, "Enable race detection")
	cmd.Flags().Bool("clone-branch", false, "Clone branch")
	cmd.Flags().String("goat-package-name", "goat", "Goat package name")
	cmd.Flags().String("goat-package-alias", "goat", "Goat package alias")
	cmd.Flags().String("goat-package-path", "goat", "Goat package path")
	cmd.Flags().String("track-strategy", "project", "Main package track strategy (project, package)")
	cmd.Flags().String("ignores", "", "Comma-separated list of files/dirs to ignore")
	cmd.Flags().String("main-entries", "", "Comma-separated list of main entries to track")

	return cmd
}
