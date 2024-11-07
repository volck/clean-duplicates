/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"clean-duplicates/internal"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var listOpt bool

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if listOpt {
			writer := internal.NewWriter(nil)
			duplicates := writer.GetDuplicates()
			duplicateInfoMsg := fmt.Sprintf("%s has %d duplicates in its database", internal.AppName, len(duplicates))
			internal.Logger.Info(duplicateInfoMsg)
			for _, duplicate := range duplicates {
				internal.Logger.Info("found duplicate", slog.Any("path", duplicate.FilePath), slog.Any("hash", duplicate.Hash))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.PersistentFlags().BoolVarP(&listOpt, "list", "l", false, "list duplicate values")
}
