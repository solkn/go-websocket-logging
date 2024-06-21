package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var messageFormat string

func main() {
	origin := "http://localhost:8080"
	url := "ws://localhost:8080/ws"

	fmt.Println("Choose message format (json or text): ")
	fmt.Scanln(&messageFormat)

	for {
		log.Println("Attempting to connect to the server")
		var conn *websocket.Conn
		var err error
		conn, _, err = websocket.DefaultDialer.Dial(url, http.Header{"Origin": {origin}})
		if err != nil {
			log.Println("Dial error:", err)
			time.Sleep(2 * time.Second)
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
				if err := sendMessage(conn, userInput); err != nil {
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

func sendMessage(conn *websocket.Conn, message string) error {
	if messageFormat == "json" {

		var data map[string]interface{}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return conn.WriteMessage(websocket.TextMessage, jsonData)
	} else if messageFormat == "text" {
		return conn.WriteMessage(websocket.TextMessage, []byte(message))
	} else {
		return fmt.Errorf("Invalid message format: %s", messageFormat)
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

		if messageFormat == "json" {
			var data map[string]interface{}
			err := json.Unmarshal(message, &data)
			if err != nil {
				fmt.Println("Error decoding JSON:", err)
				continue
			}
			fmt.Println("Received message (JSON):", data)
		} else if messageFormat == "text" {
			fmt.Printf("Received message: %s\n", message)
		} else {
			fmt.Println("Invalid message format:", messageFormat)
		}
	}
}
