package server
import (
	"bufio"
	"fmt"
	"go-websocket-logging/RTLogger"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
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
	conn        *websocket.Conn
	id          string
	lastSentPos int64
}

var clients = map[string]*client{}
var clientsMutex = &sync.Mutex{}
var fileMutex = &sync.Mutex{}
var lastReadPos int64

func generateClientID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	return id.String()
}

func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	clientID := generateClientID()
	newClient := &client{conn: conn, id: clientID}
	clientsMutex.Lock()
	clients[clientID] = newClient
	clientsMutex.Unlock()

	log.Printf("Client %s connected\n", clientID)

	logFile := "go-websocket-logging.log"

	ctr, _ := countLines(logFile)
	RTLogger.InitRTLogger(logFile, ctr)

	go KeepWritinglog()
	go watchLogFile(logFile)

	sendLogFileInfos(newClient, logFile)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				log.Printf("Client %s disconnected\n", clientID)
				clientsMutex.Lock()
				delete(clients, clientID)
				clientsMutex.Unlock()
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

func sendLogFileInfos(client *client, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()

	fileMutex.Lock()
	defer fileMutex.Unlock()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "\n" || strings.HasPrefix(line, "#") {
			continue
		}

		err := client.conn.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			log.Printf("Error sending log message to client: %v", err)
			client.conn.Close()
			return
		}
		client.lastSentPos, _ = file.Seek(0, io.SeekCurrent)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning log file: %v", err)
	}

	fmt.Println("Sent all messages from log file")
}

func KeepWritinglog() {
	for i := 0; i < 1000; i++ {
		time.Sleep(10 * time.Second)
		logFile := "go-websocket-logging.log"

		ctr, _ := countLines(logFile)
		RTLogger.InitRTLogger(logFile, ctr)
	}
}

func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

func watchLogFile(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatalf("Failed to watch file %s: %v", filePath, err)
	}

	log.Printf("Started watching file: %s\n", filePath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("event1.", event)
				log.Println("Watcher closed or program exiting. Stopping monitoring.")
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("event2.", event)
				log.Println("File modified. Notifying clients.")
				debouncedNotifyClients(filePath)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("Watcher errors channel closed. Stopping monitoring.")
				return
			}
			log.Println("Error watching file:", err)
		}
	}
}

func debouncedNotifyClients(filePath string) {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	time.Sleep(10 * time.Second)
	notifyClients(filePath)
}

func notifyClients(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()

	fileMutex.Lock()
	defer fileMutex.Unlock()

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range clients {
		file.Seek(client.lastSentPos, io.SeekStart)
		scanner := bufio.NewScanner(file)
		var buffer []string

		for scanner.Scan() {
			line := scanner.Text()
			buffer = append(buffer, line)
			client.lastSentPos, _ = file.Seek(0, io.SeekCurrent)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error scanning log file: %v", err)
		}

		for _, line := range buffer {
			if err := client.conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				log.Printf("Error sending log message to client: %v", err)
				client.conn.Close()
				delete(clients, client.id)
				break
			}
		}

		buffer = buffer[:0]
	}
}
