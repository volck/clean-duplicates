package internal

import (
	"fmt"
	"log/slog"
)

type Calculator struct {
	CalculateChan chan string
	Writer        Writer
}

func NewCalculator(thechan chan string) *Calculator {
	return &Calculator{CalculateChan: thechan}
}

func (c *Calculator) Listen() {
	Logger.Info("calculator listening")
	for {

		select {
		case path := <-c.CalculateChan:
			Logger.Info("received path", slog.Any("path", path))
			Wg.Add(1)
			go c.CalculateHash(path)
		default:
			Logger.Info("no path received")
		}

	}
}

func (c *Calculator) CalculateHash(path string) {
	Logger.Info("calculating hash", slog.Any("path", path))
	theHash, err := calculateHash(path)
	if err != nil {
		Logger.Error("could not calculate hashes", slog.Any("error", err))
	}
	md5Hash := fmt.Sprintf("%x", *theHash)
	f := File{FilePath: path, MD5Hash: md5Hash}
	Logger.Info("calculated hash", slog.Any("hash", md5Hash))
	c.Writer.WriteChan <- f

}
