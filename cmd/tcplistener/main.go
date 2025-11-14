package main

import (
	"fmt"
	"log"
	"net"

	"github.com/spaghetti-lover/go-http/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatal("Error listening to port: ", err)
	}
	defer listener.Close()

	log.Println("Server listening on port", ": 42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := request.FromReader(conn)
	if err != nil {
		log.Printf("Error reading from %s: %v", conn.RemoteAddr(), err)
		return
	}

	fmt.Print(req.String())
}
