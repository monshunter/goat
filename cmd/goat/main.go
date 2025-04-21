package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/monshunter/goat/pkg/log"
	"github.com/spf13/cobra"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "goat",
		Short: "Goat is a tool for analyzing and instrumenting Go programs",
		Long:  `Goat is a tool for analyzing and instrumenting Go programs`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set the verbose mode for the log
			log.SetVerbose(verbose)

			// Skip project checks for help or version commands
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}

			// check current directory is a golang project
			if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
				return fmt.Errorf("current directory is not a golang project")
			}
			// check current directory is a git repository
			if _, err := os.Stat(".git"); os.IsNotExist(err) {
				return fmt.Errorf("current directory is not a git repository")
			}

			log.Infof("start to run %s command", cmd.Name())
			return nil
		},
	}

	// Set the number of CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Add global persistent flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(patchCmd())
	rootCmd.AddCommand(fixCmd())
	rootCmd.AddCommand(cleanCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute root command: %v", err)
	}
}
