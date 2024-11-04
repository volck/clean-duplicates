package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type Dispatcher struct {
	writer     *Writer
	calculator *Calculator
}

func NewDispatcher(writer Writer, calculator Calculator) *Dispatcher {
	return &Dispatcher{writer: &writer, calculator: &calculator}

}

func (d *Dispatcher) FindFiles(directory string) {
	err := filepath.WalkDir(directory, d.dispatchToCalculator)
	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", directory, err)
	}

}

func (d *Dispatcher) dispatchToCalculator(path string, info os.DirEntry, err error) error {

	if err != nil {
		fmt.Printf("Error accessing path %q: %v\n", path, err)
		return err
	}
	if !info.IsDir() {
		d.calculator.CalculateChan <- path
	}
	return nil
}
