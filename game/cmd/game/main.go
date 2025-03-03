package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/internal/config"
	"github.com/ahmetkoprulu/rtrp/game/internal/websocket"
)

func main() {
	config := config.LoadEnvironment()
	fmt.Println(config)

	utils.InitLogger()
	defer utils.Logger.Sync()

	utils.SetJWTSecret(config.JWTSecret)

	wsServer := websocket.NewServer()

	go wsServer.Run()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", wsServer.HandleWebSocket)
	log.Println("Starting game server on :8080...")
	log.Println("Default room created and ready for connections")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
