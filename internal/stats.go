package internal

import (
	"net/http"
	"runtime"
	"sync"
	"time"
)

var FilesFoundChan int
var FilesProcessedChan int
var RunStart = time.Now()

type Stats struct {
	cache *map[string]bool
}

func NewStats() *Stats {
	return &Stats{cache: &map[string]bool{}}
}

func (s *Stats) statsHandler(w http.ResponseWriter, req *http.Request) {
	stats := map[string]interface{}{"GoRoutines": runtime.NumGoroutine(), "FilesFound": FilesFound, "FilesProcessed": FilesProcessed, "TimeElapsed": time.Since(RunStart).Seconds(), "cachedFiles": len(*s.cache)}
	JSON(w, http.StatusFound, stats)
}

func (s *Stats) Servestats(wg *sync.WaitGroup) {
	Logger.Info("serving stats")
	http.HandleFunc("/stats", s.statsHandler)
	http.ListenAndServe(":8080", nil)
}
