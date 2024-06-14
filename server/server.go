package main

import (
	"bufio"
	"encoding/json"
	"go-websocket-logging/logger"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	conn *websocket.Conn
	id   string
}

var clients = map[string]*client{}

func generateClientID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	return id.String()
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	clientID := generateClientID()
	newClient := &client{conn: conn, id: clientID}
	clients[clientID] = newClient

	log.Printf("Client %s connected\n", clientID)

	welcomeMessage := []byte("Welcome client!")
	if err := conn.WriteMessage(websocket.TextMessage, welcomeMessage); err != nil {
		log.Println("Error sending welcome message:", err)
	}

	sendLogFileInfos(conn)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				log.Printf("Client %s disconnected\n", clientID)
				delete(clients, clientID)
				break
			} else { 
				log.Printf("Client %s disconnected\n", clientID)

				break
			}
		}
		handleMessages(newClient, clientID, messageType, message)
	}

	conn.Close()

}

func handleMessages(c *client, clientID string, messageType int, message []byte) {
	// This function handles messages from clients, but you can leave it empty for now
}

func sendLogFileInfos(conn *websocket.Conn) {

	logFile := "go-websocket-logging.log"

	file, err := os.Open(logFile)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	sentMessages := make(map[string]struct{})

	for scanner.Scan() {
		message := scanner.Text()

		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(message), &jsonData); err == nil {
			if messageValue, ok := jsonData["message"]; ok {
				message, ok := messageValue.(string)
				if ok {
					if _, exists := sentMessages[message]; !exists {
						if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
							log.Printf("Error sending log message to client: %v", err)
							return
						}
						sentMessages[message] = struct{}{}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning log file: %v", err)
	}

}

func main() {

	logFile := "go-websocket-logging.log"

	logger.InitLogger(logFile)

	router := gin.Default()
	router.GET("/ws", wsHandler)

	log.Println("Server started on :8080")
	log.Fatal(router.Run(":8080"))
}
