package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spaghetti-lover/go-http/internal/utils"
)

func main() {
	filename := "message.txt"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	lineChannel, errorChannel := utils.GetLinesChannel(f)

	for {
		select {
		case line, ok := <-lineChannel:
			if !ok {
				return //Channel is closed
			}
			fmt.Printf("read: %s\n", line)

		case err := <-errorChannel:
			if err != nil {
				log.Fatal("Error reading file: ", err)
			}
		}
	}
}
