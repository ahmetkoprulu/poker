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

	gameManager := game.NewGameManager()
	messageHandler := websocket.NewMessageHandler(gameManager)
	wsServer := websocket.NewServer(messageHandler)

	messageHandler.SetServer(wsServer)
	go wsServer.Run()

	http.HandleFunc("/ws", wsServer.HandleWebSocket)
	log.Println("Starting game server on :8080...")
	log.Println("Default room created and ready for connections")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
