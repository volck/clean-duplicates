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

		for _, path := range paths {
			internal.FindFiles(path)
		}

		internal.Logger.Info("done finding files")
		var Files = make(chan internal.File, internal.NumberOfFiles)

		internal.Logger.Info("found files", slog.Any("length of files", len(internal.PathsFound)))

		db, err := internal.OpenDb()
		if err != nil {
			internal.Logger.Info("error opening db", slog.Any("err", err))
		}
		internal.Logger.Info("db stats", slog.Int("inuse", db.Stats().InUse), slog.Any("ping", db.Ping()))

		for _, f := range internal.PathsFound {
			internal.Wg.Add(1)
			go internal.CalculateFile(f, db)
		}

		close(Files)

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
