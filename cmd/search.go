/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"clean-duplicates/internal"
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"sync"
	"time"
)

var (
	paths       []string
	ntfy        bool
	initNewDb   bool
	enableStats bool
	enableCache bool
	NumChannels int
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use: "search",

	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		internal.Logger.Info("paths", slog.Any("paths", paths))
		if ntfy {
			ntfyStartMsg := fmt.Sprintf("%s started on %s", internal.AppName, internal.RunStart.Format("02-01-2006 at 15:04:05"))
			internal.Ntfy("clean-duplicates started search", ntfyStartMsg)
		}

		var wg sync.WaitGroup
		writerChan := make(chan internal.File, 1)
		writer := internal.NewWriter(writerChan)

		s := internal.NewStats()
		writer.InitDB()
		cache := make(map[string]bool)
		if enableCache {
			internal.Logger.Info("enabling cache")
			cache = s.FillCache(*writer)
		}
		calculateChan := make(chan string, NumChannels)
		calculator := internal.NewCalculator(calculateChan)

		dispatcher := internal.NewDispatcher(*writer, *calculator)

		if enableStats {
			internal.Logger.Info("enabling stats")
			wg.Add(1)
			go s.Servestats(&wg)
		}

		wg.Add(1)
		go calculator.Listen(calculateChan, writerChan, cache, &wg)
		wg.Add(1)
		go writer.Listen(writerChan, &wg)

		dispatcher.FindFiles(paths)
		wg.Wait()

		after := time.Since(internal.RunStart)
		internal.Logger.Info("clean-duplicates finished", slog.Any("runtime", after.Seconds()), slog.Any("files", internal.FilesFound), slog.Any("processed", internal.FilesProcessed), slog.Any("cached", len(cache)), slog.Any("NumberofChannels", NumChannels))
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
	searchCmd.PersistentFlags().IntVar(&NumChannels, "numChannels", 4, "number of channels")
	searchCmd.PersistentFlags().BoolVar(&initNewDb, "initDb", false, "initialize new db")
	searchCmd.PersistentFlags().BoolVar(&enableStats, "stats", false, "enable webserver and influxdb stats")
	searchCmd.PersistentFlags().BoolVar(&enableCache, "cache", true, "enables cache based on database entries")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
