/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"clean-duplicates/internal"
	"fmt"
	"log/slog"
	"time"

	//	"time"

	"github.com/spf13/cobra"
)

var (
	paths []string
	ntfy  bool
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use: "search",

	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		internal.Logger.Info("paths", slog.Any("paths", paths))
		t := time.Now()
		if ntfy {
			ntfyStartMsg := fmt.Sprintf("%s started on %s", internal.AppName, t.Format("02-01-2006 at 15:04:05"))
			internal.Ntfy("clean-duplicates started search", ntfyStartMsg)
		}

		writerChan := make(chan internal.File, 20)
		writer := internal.NewWriter(writerChan)

		calculateChan := make(chan string, 100)
		calculator := internal.NewCalculator(calculateChan)

		dispatcher := internal.NewDispatcher(*writer, *calculator)

		for _, path := range paths {

			dispatcher.FindFiles(path)

		}

		after := time.Since(t)
		internal.Logger.Info("clean-duplicates finished", slog.Any("runtime", after))
		if ntfy {
			ntfyTitle := fmt.Sprintf("%s completed search for files", internal.AppName)
			tnow := time.Now()
			ntfyMsg := fmt.Sprintf("%s completed in %s at %s ", internal.AppName, after, tnow.Format("02-01-2006 at 15:04:05"))
			internal.Ntfy(ntfyTitle, ntfyMsg)
		}

	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	searchCmd.PersistentFlags().StringArrayVar(&paths, "path", []string{}, "define paths")
	searchCmd.PersistentFlags().BoolVarP(&ntfy, "notify", "n", false, "toggle ntfy")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
