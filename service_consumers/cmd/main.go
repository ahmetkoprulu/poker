// cmd/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ahmetkoprulu/rtrp/consumers/common/data"
	"github.com/ahmetkoprulu/rtrp/consumers/common/utils"
	"github.com/ahmetkoprulu/rtrp/consumers/internal"
	"github.com/ahmetkoprulu/rtrp/consumers/internal/config"
	"github.com/ahmetkoprulu/rtrp/consumers/internal/consumers"
	"github.com/ahmetkoprulu/rtrp/consumers/internal/mq"
	"github.com/ahmetkoprulu/rtrp/consumers/models"
)

func main() {
	cfg := config.LoadEnvironment()
	fmt.Println(cfg)

	utils.InitLogger()
	defer utils.Logger.Sync()

	utils.SetJWTSecret(cfg.JWTSecret)
	err := data.LoadPostgres(cfg.DatabaseURL, cfg.DatabaseName, false)
	if err != nil {
		log.Fatalf("Failed to load Postgres: %v\n", err)
	}

	db, err := data.NewPgDbContext()
	if err != nil {
		utils.Logger.Fatal("Failed to connect to database", utils.Logger.String("error", err.Error()))
	}
	defer db.Close()

	_, err = initMq(cfg)
	if err != nil {
		log.Fatal("Failed to initialize MQ: ", err)
	}

	consumers := map[string]consumers.IConsumer{
		"chip-update": consumers.NewChipUpdateConsumer("chip-update", db),
	}

	consumerManager, err := internal.NewConsumerManager(db, consumers)
	if err != nil {
		log.Fatal("Failed to create consumer service:", err)
	}

	consumerManager.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down consumer service...")
	consumerManager.Shutdown()
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
