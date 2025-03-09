package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/internal"
	"github.com/ahmetkoprulu/rtrp/game/internal/config"
)

func main() {
	config := config.LoadEnvironment()
	fmt.Println(config)

	utils.InitLogger()
	defer utils.Logger.Sync()

	utils.SetJWTSecret(config.JWTSecret)

	wsServer := internal.NewServer()

	go wsServer.Run()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", wsServer.HandleWebSocket)
	log.Println("Starting game server on :" + config.ServerPort + "...")
	log.Println("Default room created and ready for connections")
	if err := http.ListenAndServe(":"+config.ServerPort, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
