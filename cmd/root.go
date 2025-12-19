package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "specture",
	Short: "A spec-driven software architecture system",
	Long:  `Specture is a spec-driven software architecture system. It provides a lightweight, document-driven approach to project planning.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(validateCmd)
}
