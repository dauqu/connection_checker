package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// connectedClients is a map that stores connected clients.
	connectedClients = make(map[*websocket.Conn]bool)
	// broadcast is a channel for sending messages to all connected clients.
	broadcast = make(chan string)
	// upgrader is used to upgrade an HTTP connection to a WebSocket connection.
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	// Start a background goroutine to listen for broadcast messages and send them to all connected clients.
	go handleBroadcast()

	// Create an HTTP server and register a WebSocket handler.
	http.HandleFunc("/ws", handleWebSocket)

	// Start the server and listen for connections.
	fmt.Println("Starting server...")
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// handleWebSocket is a handler function for WebSocket connections.
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Add the client to the connectedClients map.
	connectedClients[conn] = true

	// Send a welcome message to the client.
	err = conn.WriteMessage(websocket.TextMessage, []byte("Welcome!"))
	if err != nil {
		log.Println(err)
	}

	// Start a background goroutine to listen for messages from the client.
	go handleWebSocketMessages(conn)
}

// handleWebSocketMessages listens for messages from a WebSocket client and sends them to all other connected clients.
func handleWebSocketMessages(conn *websocket.Conn) {
	defer func() {
		// When this function returns, remove the client from the connectedClients map and close the connection.
		delete(connectedClients, conn)
		conn.Close()
	}()

	for {
		// Read a message from the client.
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Send the message to all other connected clients.
		broadcast <- fmt.Sprintf("[%s]: %s", time.Now().Format("2006-01-02 15:04:05"), string(message))
	}
}

// handleBroadcast listens for messages on the broadcast channel and sends them to all connected clients.
func handleBroadcast() {
	for {
		// Wait for a message on the broadcast channel.
		message := <-broadcast

		// Send the message to all connected clients.
		for conn := range connectedClients {
			err := conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println(err)
			}
		}

		// Print the current list of connected clients in the terminal.
		// fmt.Println("Connected clients:")
		// for conn := range connectedClients {
		// 	fmt.Println(conn)
		// 	//Print the clients length
		// 	fmt.Println(len(connectedClients))
		// }
		
		fmt.Println(len(connectedClients))
	}
}
