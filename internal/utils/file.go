package utils

import (
	"bufio"
	"io"
	"log"
)

func GetLinesChannel(f io.ReadCloser) (<-chan string, <-chan error) {
	lineChannel := make(chan string, 1)
	errorChannel := make(chan error, 1)

	go func() {
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				select {
				case errorChannel <- closeErr:
				default:
					log.Printf("Failed to close file: %v", closeErr)
				}
			}
		}()
		defer close(lineChannel)
		defer close(errorChannel)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			lineChannel <- line
		}

		if err := scanner.Err(); err != nil {
			errorChannel <- err
		}
	}()

	return lineChannel, errorChannel
}
