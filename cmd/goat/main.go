package main

import (
	"fmt"
	"os"

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
			// 设置日志的verbose模式
			log.SetVerbose(verbose)

			// 如果是help命令或version命令，则跳过项目检查
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

	// 添加全局持久化标志
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(patchCmd())
	rootCmd.AddCommand(fixCmd())
	rootCmd.AddCommand(cleanCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute root command: %v", err)
	}
}
