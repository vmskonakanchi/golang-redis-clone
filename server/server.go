package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	DEFAULT_PORT        = "6969"
	MAX_BUFFER_SIZE     = 1024
	DEFAULT_HOST        = "localhost"
	COMMAND_SET         = "SET"
	COMMAND_GET         = "GET"
	COMMAND_DEL         = "DEL"
	COMMAND_NOTIFY      = "NOTIFY"
	COMMAND_PING        = "PING"
	COMMAND_PONG        = "PONG"
	COMMAND_QUIT        = "QUIT"
	COMMAND_KEYS        = "KEYS"
	COMMAND_ADD_REPLICA = "ADDREPLICA"
)

type Entry struct {
	Key      string
	Value    string
	DataType string // "string", "number", "json"
}

var clients map[net.Addr]net.Conn = make(map[net.Addr]net.Conn, 100)
var entries map[string]Entry = make(map[string]Entry)
var notifications map[string][]net.Conn = make(map[string][]net.Conn)
var replicas map[string]net.Conn = make(map[string]net.Conn, 10)

func main() {
	port := DEFAULT_PORT
	host := DEFAULT_HOST

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))

	if err != nil {
		log.Fatalln("Error starting server ", err)
		return
	}

	log.Println("Server started on port", port)

	changes := make(chan []byte)

	go doReplication(changes)

	for {
		client, err := conn.Accept()

		if err != nil {
			log.Fatalln("Error starting server", err)
			continue
		}

		go handleClient(client, changes)
	}
}

func handleClient(conn net.Conn, changes chan []byte) {
	// Log the connection details
	clients[conn.RemoteAddr()] = conn
	log.Println("Client connected from ", conn.RemoteAddr())
	defer delete(clients, conn.RemoteAddr())
	defer log.Println("Client disconnected from ", conn.RemoteAddr())
	defer conn.Close()

	for {
		buffer := make([]byte, MAX_BUFFER_SIZE)
		n, err := conn.Read(buffer)

		if err != nil {
			log.Println("Error reading from client: ", err)
			return
		}

		if n == 0 {
			log.Println("Client closed connection")
			return
		}

		if uint16(n) > MAX_BUFFER_SIZE {
			log.Println("Received message exceeds buffer size, truncating")
			n = int(MAX_BUFFER_SIZE)
		}

		msg := string(buffer[:n])

		splittedMsg := strings.Split(msg, " ")
		action := splittedMsg[0]

		log.Printf("Received command: %s from %s", action, conn.RemoteAddr())

		switch action {
		case COMMAND_PING:
			go handlePing(conn)
		case COMMAND_SET:
			go handleSet(conn, splittedMsg)
		case COMMAND_GET:
			go handleGet(conn, splittedMsg)
		case COMMAND_DEL:
			go handleDel(conn, splittedMsg)
		case COMMAND_KEYS:
			go handleGetKeys(conn)
		case COMMAND_NOTIFY:
			go handleNotify(conn, splittedMsg)
		case COMMAND_ADD_REPLICA:
			go handleAddReplica(conn, splittedMsg)
		default:
			log.Println("Error unknown command: ", action)
		}

		// add this command to the changes channel for replication
		if action == COMMAND_SET || action == COMMAND_DEL {
			changes <- buffer
		}

	}
}

func handlePing(conn net.Conn) {
	_, err := conn.Write([]byte(COMMAND_PONG))
	if err != nil {
		log.Println("Error sending PONG response: ", err)
		return
	}
	log.Println("Sent PONG response")
}

func handleSet(conn net.Conn, args []string) {
	if len(args) < 3 {
		log.Println("Error: SET command requires a key and a value")
		_, err := conn.Write([]byte("Error: SET command requires a key and a value"))
		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	key := args[1]
	value := strings.Join(args[2:], " ")
	var dataType string

	switch value[0] {
	case '{', '[', '"':
		dataType = "json"
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		dataType = "number"
	default:
		dataType = "string"
	}

	// Check if the key already exists
	if existingEntry, exists := entries[key]; exists {
		log.Printf("Key %s already exists, updating value", key)
		// Notify all clients subscribed to this key
		if existingEntry.Value != value {
			if clients, ok := notifications[key]; ok {
				for _, client := range clients {

					if client == conn { // Don't notify the client that made the change
						continue
					}

					_, err := fmt.Fprintf(client, "Key %s updated to %s", key, value)

					if err != nil {
						log.Printf("Error notifying client %s: %v", client.RemoteAddr(), err)
					}
				}
			}
		}
	}

	// Create or update the entry
	entries[key] = Entry{
		Key:      key,
		Value:    value,
		DataType: dataType,
	}

	_, err := fmt.Fprintf(conn, "OK")

	if err != nil {
		log.Println("Error sending response: ", err)
		return
	}
}

func handleGet(conn net.Conn, args []string) {
	if len(args) < 2 {
		log.Println("Error: GET command requires a key")

		_, err := fmt.Fprintf(conn, "Error: GET command requires a key")

		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	key := args[1]
	entry, exists := entries[key]
	if !exists {
		log.Printf("Key %s not found", key)
		_, err := fmt.Fprintf(conn, "Key not found")
		if err != nil {
			log.Println("Error sending response: ", err)
		}
		return
	}
	fmt.Fprintf(conn, "%s", entry.Value)
}

func handleDel(conn net.Conn, args []string) {
	if len(args) < 2 {
		log.Println("Error: DEL command requires a key")
		_, err := fmt.Fprintf(conn, "Error: DEL command requires a key")
		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	key := args[1]
	if _, exists := entries[key]; !exists {
		log.Printf("Key %s not found", key)
		_, err := fmt.Fprintf(conn, "Key not found")
		if err != nil {
			log.Println("Error sending response: ", err)
		}
		return
	}

	delete(entries, key)
	fmt.Fprintf(conn, "OK")
}

func handleGetKeys(conn net.Conn) {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	response := strings.Join(keys, "\n")
	if response == "" {
		response = "No keys found"
	}
	_, err := fmt.Fprintf(conn, "%s", response)
	if err != nil {
		log.Println("Error sending KEYS response: ", err)
		return
	}
	log.Printf("Sent KEYS response: %s", response)
}

func handleNotify(conn net.Conn, args []string) {
	if len(args) < 2 {
		log.Println("Error: NOTIFY command requires a key to subscribe to")
		_, err := fmt.Fprintf(conn, "Error: NOTIFY command requires a key to subscribe to")
		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	key := args[1]

	// check if the key exists in the entries map
	if _, exists := entries[key]; !exists {
		log.Printf("Key %s does not exist, cannot subscribe", key)
		_, err := fmt.Fprintf(conn, "Key %s does not exist, cannot subscribe", key)
		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	// add this client to the list of clients for this key
	if _, exists := notifications[key]; !exists {
		notifications[key] = make([]net.Conn, 0)
	}

	notifications[key] = append(notifications[key], conn)

	log.Printf("Client subscribed to key %s", key)
}

func doReplication(changes <-chan []byte) {
	for change := range changes {
		if len(replicas) > 0 {
			for _, replicaConn := range replicas {
				// sending the command to the replica
				_, err := replicaConn.Write(change)

				if err != nil {
					log.Printf("Error sending command to replica %s: %v", replicaConn.RemoteAddr(), err)
					continue
				}
			}
		} else {
			log.Println("No replicas available to replicate changes")
			continue
		}
	}
}

func handleAddReplica(conn net.Conn, args []string) {
	if len(args) < 2 {
		log.Println("Error: ADDREPLICA command requires a host and port")
		_, err := fmt.Fprintf(conn, "Error: ADDREPLICA command requires a host and port")
		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	hostPort := args[1]
	replicaConn, err := net.Dial("tcp", hostPort)
	if err != nil {
		log.Printf("Error connecting to replica %s: %v", hostPort, err)
		_, err := fmt.Fprintf(conn, "Error connecting to replica %s", hostPort)
		if err != nil {
			log.Println("Error sending error response: ", err)
		}
		return
	}

	if _, exists := replicas[hostPort]; exists {
		fmt.Fprintf(conn, "Replica at %s already exists", hostPort)
		log.Printf("Replica at %s already exists", hostPort)
		return
	}

	replicas[hostPort] = replicaConn
	log.Printf("Added replica at %s", hostPort)
}
