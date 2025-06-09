package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/internal"
	"github.com/ahmetkoprulu/rtrp/game/internal/config"
	"github.com/ahmetkoprulu/rtrp/game/internal/mq"
	"github.com/ahmetkoprulu/rtrp/game/models"
)

func main() {
	config := config.LoadEnvironment()
	fmt.Println(config)

	utils.InitLogger()
	defer utils.Logger.Sync()

	utils.SetJWTSecret(config.JWTSecret)

	_, err := initMq(config)
	if err != nil {
		log.Fatal("Failed to initialize MQ: ", err)
	}

	wsServer := internal.NewServer()

	go wsServer.Run()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", wsServer.HandleWebSocket)
	http.HandleFunc("/rooms", wsServer.HandleRoomList)
	http.HandleFunc("/admin/reset", wsServer.HandleReset)
	log.Println("Starting game server on :" + config.ServerPort + "...")
	log.Println("Default room created and ready for connections")
	if err := http.ListenAndServe(":"+config.ServerPort, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func initMq(config *models.Config) (*mq.MqClient, error) {
	mq.InitMq(config.MqURL)

	mqProvider, err := mq.NewMqClient()
	if err != nil {
		return nil, err
	}

	mqProvider.Provider.DeclareExchange(mq.GameExchange, "topic", true)
	// mqProvider.Provider.DeclareQueue(mq.ChipUpdatesQueue, true, mq.ChipUpdatesRoutingKey, mq.GameExchange)
	// mqProvider.Provider.DeclareQueue(mq.GameAnalyticsQueue, true, mq.GameAnalyticsRoutingKey, mq.GameExchange)

	return mqProvider, nil
}
