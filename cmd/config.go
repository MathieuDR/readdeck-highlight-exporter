/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	// "fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate or view your configuration",
	// LEARNING: Cannot escape ` in raw strings. [link](https://github.com/golang/go/issues/24475)
	// LEARNING: You need to have a `,` at the end, even if it's the last thing in this object/struct/class(?)
}

func init() {
	rootCmd.AddCommand(configCmd)
}
