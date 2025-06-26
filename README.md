# Go Redis Clone

A lightweight, educational implementation of Redis written in Go, featuring core Redis functionality including key-value storage, replication, and real-time notifications.

## Overview

This project is a simplified Redis clone that demonstrates fundamental distributed systems concepts including:

- TCP-based client-server architecture
- Key-value storage with multiple data types
- Master-slave replication
- Real-time change notifications
- Concurrent client handling

## Features

### Core Redis Commands

- **SET key value** - Store a key-value pair
- **GET key** - Retrieve value by key
- **DEL key** - Delete a key-value pair
- **KEYS** - List all stored keys
- **PING** - Health check (returns PONG)

### Advanced Features

- **NOTIFY key** - Subscribe to key change notifications
- **ADDREPLICA host:port** - Add a replica server for data replication
- **Data Type Detection** - Automatic detection of string, number, and JSON data types
- **Real-time Replication** - Changes are automatically propagated to replica servers
- **Concurrent Client Support** - Multiple clients can connect simultaneously

## Architecture

### Server Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Client 1    │    │     Client 2    │    │     Client N    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │      Redis Server       │
                    │    (localhost:6969)     │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │    Replication Layer    │
                    └────────────┬────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
│    Replica 1    │    │    Replica 2    │    │    Replica N    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components

1. **Server (`server/server.go`)**

   - Main server logic handling TCP connections
   - Command parsing and routing
   - In-memory data storage
   - Replication management
2. **Client (`client/client.go`)**

   - Interactive command-line client
   - Concurrent read/write operations
   - Connection management
3. **Data Structures**

   - `Entry`: Stores key-value pairs with data type metadata
   - `clients`: Map of connected clients
   - `entries`: In-memory key-value store
   - `notifications`: Subscription management for key changes
   - `replicas`: Connected replica servers

## Installation & Usage

### Prerequisites

- Go 1.20 or higher

### Running the Server

```bash
# Start server on default port (6969)
cd server
go run server.go

# Start server on custom port
go run server.go 8080
```

### Running the Client

```bash
# Connect to default server (localhost:6969)
cd client
go run client.go

# Connect to custom server
go run client.go localhost:8080
```

### Example Usage

```bash
# Basic operations
SET name "John Doe"
GET name
DEL name

# Working with different data types
SET age 25
SET config {"host": "localhost", "port": 8080}
SET items [1, 2, 3, 4, 5]

# List all keys
KEYS

# Subscribe to key changes
NOTIFY name

# Add replica server
ADDREPLICA localhost:7000

# Health check
PING
```

## Data Types

The server automatically detects and categorizes data types:

- **String**: Default type for text values
- **Number**: Values starting with digits (0-9)
- **JSON**: Values starting with `{`, `[`, or `"`

## Replication

The server supports master-slave replication:

1. **Adding Replicas**: Use `ADDREPLICA host:port` to connect replica servers
2. **Automatic Sync**: SET and DEL operations are automatically replicated
3. **Asynchronous**: Replication happens in background goroutines
4. **Fault Tolerance**: Failed replica connections are logged and skipped

## Notifications

Real-time change notifications allow clients to subscribe to key updates:

1. **Subscribe**: `NOTIFY keyname` to receive updates for a specific key
2. **Automatic Alerts**: Subscribers receive notifications when key values change
3. **Exclude Self**: Clients don't receive notifications for their own changes

## Testing

Run the test suite to verify server functionality:

```bash
cd server
go test -v
```

Tests cover:

- Basic SET/GET/DEL operations
- Key not found scenarios
- PING/PONG health checks

## Configuration

### Server Configuration

- **Default Host**: localhost
- **Default Port**: 6969
- **Max Buffer Size**: 1024 bytes
- **Max Clients**: 100 concurrent connections
- **Max Replicas**: 10 replica servers

### Customization

Modify constants in `server/server.go`:

```go
const (
    DEFAULT_PORT        = "6969"
    MAX_BUFFER_SIZE     = 1024
    DEFAULT_HOST        = "localhost"
)
```

## Project Structure

```
go-redis-clone/
├── go.mod                 # Go module definition
├── README.md             # This documentation
├── client/
│   └── client.go         # Interactive client implementation
└── server/
    ├── server.go         # Main server implementation
    └── server_test.go    # Unit tests
```

## Limitations

This is an educational implementation with the following limitations:

- **In-Memory Only**: No persistence to disk
- **No Authentication**: No security mechanisms
- **Limited Commands**: Subset of Redis commands
- **No Clustering**: Single-node architecture with replication only
- **No TTL**: Keys don't expire
- **No Transactions**: No MULTI/EXEC support
- **Basic Error Handling**: Simplified error responses

## Contributing

This project is designed for educational purposes. Feel free to:

1. Add new Redis commands
2. Implement data persistence
3. Add authentication mechanisms
4. Improve error handling
5. Add more comprehensive tests

## License

This project is open source and available under the MIT License.

## Related Technologies

- **TCP Networking**: Low-level socket programming
- **Goroutines**: Concurrent client handling
- **Channels**: Inter-goroutine communication for replication
- **Go Standard Library**: net, strings, fmt packages
