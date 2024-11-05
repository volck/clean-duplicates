package internal_test

import (
	internal "clean-duplicates/internal"
	"sync"
	"testing"
)

func TestWorkFlow(t *testing.T) {
	t.Log("Testing the workflow")
	var wg sync.WaitGroup
	writerChan := make(chan internal.File, 20)
	writer := internal.NewWriter(writerChan)
	writer.InitDB()
	calculateChan := make(chan string, 100)
	calculator := internal.NewCalculator(calculateChan)

	dispatcher := internal.NewDispatcher(*writer, *calculator)

	paths := []string{"/home/emil/dev/scratch/clean-duplicates/testfolder/"}
	wg.Add(1)
	go calculator.Listen(calculateChan, writerChan, &wg)
	internal.Logger.Info("calculator started. we're starting writer")
	wg.Add(1)
	go writer.Listen(writerChan, &wg)
	internal.Logger.Info("writer started. we're starting dispatcher")
	calculateChan <- "test"
	writerChan <- internal.File{FilePath: "test"}
	for _, path := range paths {
		dispatcher.FindFiles(path)
	}
}
