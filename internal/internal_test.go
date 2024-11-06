package internal_test

import (
	internal "clean-duplicates/internal"
	"runtime"
	"sync"
	"testing"
)

func TestWorkFlow(t *testing.T) {
	t.Log("Testing the workflow")
	var wg sync.WaitGroup
	writerChan := make(chan internal.File, 20)
	writer := internal.NewWriter(writerChan)
	writer.DeleteDb()
	writer.InitDB()

	totalCPUs := runtime.NumCPU()
	calculateChan := make(chan string, totalCPUs)
	calculator := internal.NewCalculator(calculateChan)

	dispatcher := internal.NewDispatcher(*writer, *calculator)

	paths := []string{"/home/a01631/dev/clean-duplicates/testfolder/"}
	wg.Add(1)
	go calculator.Listen(calculateChan, writerChan, &wg)
	internal.Logger.Info("calculator started. we're starting writer")
	wg.Add(1)
	go writer.Listen(writerChan, &wg)

	dispatcher.FindFiles(paths)
	wg.Wait()
}
