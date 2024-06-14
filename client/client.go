package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	origin := "http://localhost:8080"
	url := "ws://localhost:8080/ws"

	for {
		log.Println("Attempting to connect to the server")
		var conn *websocket.Conn
		var err error
		conn, _, err = websocket.DefaultDialer.Dial(url, http.Header{"Origin": {origin}})
		if err != nil {
			log.Println("Dial error:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Connected to the server")

		defer func() {
			log.Println("Closing connection")
			conn.Close()
		}()

		go readMessages(conn)

		for {
			fmt.Print("Enter a message to send (or press Enter to exit): ")
			var userInput string
			fmt.Scanln(&userInput)

			if userInput != "" {
				log.Printf("Sending message: %s\n", userInput)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(userInput)); err != nil {
					log.Println("Error sending message:", err)
					break 
				}
			} else {
				log.Println("Exiting message loop")
				break
			}
		}
	}
}

func readMessages(conn *websocket.Conn) {
	for {
		log.Println("Waiting for message from server")
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}
		fmt.Printf("Received message: %s\n", message)
	}
}
