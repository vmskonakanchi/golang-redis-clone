package main

import (
	"bufio"
	l "log"
	"net"
	"os"
	"sync"
)

var log = l.New(os.Stdout, "", l.LstdFlags|l.Lshortfile)

func main() {
	var host string = "localhost:6969"

	if len(os.Args) > 1 {
		host = os.Args[1]
	}

	conn, err := net.Dial("tcp", host)

	if err != nil {
		log.Printf("Error connecting to the server")
		return
	}

	defer log.Printf("Disconnected from server")
	defer log.Fatal(conn.Close())

	log.Printf("Connected to the server")

	done := make(chan bool, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Start goroutines for reading and writing
	// Use a buffered channel to signal when we're done
	// This allows us to close the connection gracefully
	go writeResponse(conn, done)
	go readResponse(conn, done)

	go func() {
		if <-done {
			log.Println("Server closed connection")
			wg.Done()
		}
	}()

	wg.Wait() // Wait for both goroutines to finish
}

func readResponse(conn net.Conn, done chan bool) {
	for {
		// Read from the server
		buffer := make([]byte, 1024) // Reset the buffer for each read
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("Error reading from server: ", err)
			done <- true // Signal that we're done
			return
		}
		if n == 0 {
			log.Println("Server closed connection")
			done <- true // Signal that we're done
			return
		}
		if n > len(buffer) {
			log.Println("Received message exceeds buffer size, truncating")
			n = len(buffer)
		}
		msg := string(buffer[:n])
		log.Printf("%s", msg)
	}
}

func writeResponse(conn net.Conn, done chan bool) {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()

		_, err := conn.Write([]byte(input))

		if err != nil {
			log.Println("Error sending message to server: ", err)
			done <- true // Signal that we're done
			return
		}
	}
}
