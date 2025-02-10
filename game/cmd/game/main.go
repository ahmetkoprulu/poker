package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/internal/config"
	"github.com/ahmetkoprulu/rtrp/game/internal/game"
	"github.com/ahmetkoprulu/rtrp/game/internal/websocket"
)

func main() {
	config := config.LoadEnvironment()
	fmt.Println(config)

	utils.InitLogger()
	defer utils.Logger.Sync()

	utils.SetJWTSecret(config.JWTSecret)

	// Create game manager
	gameManager := game.NewGameManager()

	// Create message handler
	messageHandler := websocket.NewMessageHandler(gameManager)

	// Create WebSocket server
	wsServer := websocket.NewServer(messageHandler)

	// Set server reference in message handler
	messageHandler.SetServer(wsServer)

	// Start WebSocket server in a goroutine
	go wsServer.Run()

	// Set up WebSocket endpoint
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	// Start HTTP server
	log.Println("Starting game server on :8080...")
	log.Println("Default room created and ready for connections")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
