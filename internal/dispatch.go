package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Dispatcher struct {
	writer     *Writer
	dispatched int
	calculator *Calculator
}

func NewDispatcher(writer Writer, calculator Calculator) *Dispatcher {
	return &Dispatcher{writer: &writer, calculator: &calculator}

}

func (d *Dispatcher) FindFiles(paths []string) {

	Logger.Info("dispatch finding files", slog.Any("paths", paths))
	for _, path := range paths {
		err := filepath.WalkDir(path, d.dispatchToCalculator)
		if err != nil {
			fmt.Printf("Error walking the path %v: %v\n", path, err)
		}
	}
	close(d.calculator.CalculateChan)
}

func (d *Dispatcher) dispatchToCalculator(path string, info os.DirEntry, err error) error {

	if err != nil {
		fmt.Printf("Error accessing path %q: %v\n", path, err)
		return err
	}
	if !info.IsDir() {
		d.dispatched++
		Logger.Debug("dispatching to calculator", slog.Any("path", path), slog.Any("info", info), slog.Int("dispatched", d.dispatched))
		d.calculator.CalculateChan <- path
	}
	return nil
}
