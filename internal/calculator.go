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
	Logger.Info("calculator listening")
	Logger.Info("calculator sent test file")
	for {
		for ch := range c.CalculateChan {
			Logger.Info("calculator recieved path", slog.Any("path", ch))
			outChan <- File{FilePath: ch}
		}
	}
}

func (c *Calculator) CalculateHash(path string) {
	Logger.Info("calculator started", slog.Any("path", path))
	theHash, err := calculateHash(path)
	if err != nil {
		Logger.Error("could not calculate hashes", slog.Any("error", err))
	}
	md5Hash := fmt.Sprintf("%x", *theHash)
	f := File{FilePath: path, MD5Hash: md5Hash}
	Logger.Info("calculator stopped", slog.Any("hash", md5Hash))
	c.Writer.WriteChan <- f

}
