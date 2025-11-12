package utils

import (
	"bufio"
	"io"
	"log"
)

func GetLinesChannel(f io.ReadCloser) (lineChannel <-chan string, errorChannel <-chan error) {
	lineChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(lineChan)
		defer close(errChan)
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("Error closing file: %v", err)
			}
		}()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			lineChan <- line
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	return lineChan, errChan
}
