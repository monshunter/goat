package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var configYaml string

func init() {
	configYaml = ".goat.yaml"
	if os.Getenv("GOAT_CONFIG") != "" {
		configYaml = os.Getenv("GOAT_CONFIG")
	}
}
func main() {
	rootCmd := &cobra.Command{
		Use:   "goat",
		Short: "Goat is a tool for analyzing and instrumenting Go programs",
		Long:  `Goat is a tool for analyzing and instrumenting Go programs`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// check current directory is a golang project
			if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
				return fmt.Errorf("current directory is not a golang project")
			}
			// check current directory is a git repository
			if _, err := os.Stat(".git"); os.IsNotExist(err) {
				return fmt.Errorf("current directory is not a git repository")
			}
			return nil
		},
	}
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(patchCmd())
	rootCmd.AddCommand(fixCmd())
	rootCmd.AddCommand(cleanCmd())
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute root command: %v", err)
	}
}
