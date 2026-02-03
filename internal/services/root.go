package services

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper is a secure secrets manager",
	Long: `A robust and secure secrets manager that allows you to store and retrieve
your sensitive information like logins, passwords, and other private data.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Set the version string for the root command
	rootCmd.Version = fmt.Sprintf("%s (Build Date: %s)", Version, BuildDate)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add other commands here if they are not added in their own init functions
}
