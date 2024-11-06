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

func (c *Calculator) Listen(inChan chan<- string, outChan chan<- File, wg *sync.WaitGroup) {
	defer wg.Done()
	for ch := range c.CalculateChan {
		wg.Add(1)
		go c.CalculateHash(ch, outChan, wg)
	}
	Logger.Info("calculator done")
}

func (c *Calculator) CalculateHash(path string, outChan chan<- File, wg *sync.WaitGroup) {
	defer wg.Done()
	theHash, err := calculateHash(path)
	if err != nil {
		Logger.Error("could not calculate hashes", slog.Any("error", err))
	}
	md5Hash := fmt.Sprintf("%x", *theHash)
	f := File{FilePath: path, MD5Hash: md5Hash}
	outChan <- f
}
