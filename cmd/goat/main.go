package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/monshunter/goat/pkg/log"
	"github.com/spf13/cobra"
)

var verbose bool
var showVersion bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "goat",
		Short: "Goat is a tool for analyzing and instrumenting Go programs",
		Long:  `Goat is a tool for analyzing and instrumenting Go programs`,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Printf("Goat version %s\n", Version)
				return
			}
			cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set the verbose mode for the log
			log.SetVerbose(verbose)

			// Skip project checks for help or version commands
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}

			// Skip project checks when showing version
			if showVersion {
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

			log.Infof("Start to run %s command", cmd.Name())
			return nil
		},
	}

	// Set the number of CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Add global persistent flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "show version information")

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(trackCmd())
	rootCmd.AddCommand(patchCmd())
	rootCmd.AddCommand(cleanCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute root command: %v", err)
	}
}
