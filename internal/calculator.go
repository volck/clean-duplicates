package internal

import (
	"fmt"
	"log/slog"
	"sync"
)

type Calculator struct {
	CalculateChan chan string
	Writer        Writer
}

func NewCalculator(thechan chan string) *Calculator {
	return &Calculator{CalculateChan: thechan}
}

func (c *Calculator) Listen(inChan chan<- string, outChan chan<- File, cache map[string]bool, wg *sync.WaitGroup) {
	defer wg.Done()
	Logger.Info("calculator listening")
	var calculateWg sync.WaitGroup
	for ch := range c.CalculateChan {
		if _, ok := cache[ch]; ok {
			Logger.Info("skipping", slog.Any("path", ch), slog.Any("hash", cache[ch]))
			continue
		}
		calculateWg.Add(1)
		go c.CalculateHash(ch, outChan, &calculateWg)
	}
	calculateWg.Wait()
	close(outChan)
	wg.Done()
}

func (c *Calculator) CalculateHash(path string, outChan chan<- File, wg *sync.WaitGroup) {
	defer wg.Done()
	theHash, err := calculateHash(path)
	if err != nil {
		Logger.Error("could not calculate hashes", slog.Any("error", err))
		return
	}
	md5Hash := fmt.Sprintf("%x", *theHash)
	f := File{FilePath: path, Hash: md5Hash}
	FilesProcessed = FilesProcessed + 1
	outChan <- f
}
