package internal_test

import (
	internal "clean-duplicates/internal"
	"testing"
)

func TestWorkFlow(t *testing.T) {
	t.Log("Testing the workflow")
	writerChan := make(chan internal.File, 20)
	writer := internal.NewWriter(writerChan)
	writer.InitDB()
	calculateChan := make(chan string, 100)
	calculator := internal.NewCalculator(calculateChan)

	dispatcher := internal.NewDispatcher(*writer, *calculator)

	paths := []string{"/home/a01631/dev/clean-duplicates/testfolder"}

	go calculator.Listen()
	go writer.Listen()

	for _, path := range paths {
		dispatcher.FindFiles(path)
	}
}
