package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/vimalraj-a/mongo-hack/server/internal/config"
	"github.com/vimalraj-a/mongo-hack/server/internal/database"
	"github.com/vimalraj-a/mongo-hack/server/internal/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	cfg := config.Load()

	mongoClient, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	defer func() {
		if err := database.Disconnect(mongoClient); err != nil {
			log.Printf("error disconnecting mongo: %v", err)
		}
	}()

	r := router.New(mongoClient, cfg)

	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
