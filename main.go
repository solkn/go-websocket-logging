package main

import (
	"go-websocket-logging/server"
	"log"

	"github.com/gin-gonic/gin"
)
func main() {

	router := gin.Default()
	router.GET("/ws", server.WsHandler)

	log.Println("Server started on :8080")
	log.Fatal(router.Run(":8080"))
}