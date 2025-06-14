package main

import (
	"context"
	"log"
	"qaBot/internal/bot"
	"qaBot/internal/infrastructure/database"
	"qaBot/pkg/config"
	"time"
)

func main() {
	// Retrieve configuration values
	botToken := config.GetString("bot_token")
	workers := config.GetInt("workers")
	if workers <= 0 {
		log.Fatal("Invalid number of workers specified in configuration")
		workers = 1 // Default to 1 worker if not specified or invalid
	}

	pgCfg := database.PostgresConfig{
		Host:            config.GetString("database.host"),
		Port:            config.GetString("database.port"),
		User:            config.GetString("database.user"),
		Password:        config.GetString("database.password"),
		DBName:          config.GetString("database.name"),
		SSLMode:         config.GetString("database.sslmode"),
		MaxConns:        int32(config.GetInt("database.max_conns")),
		MinConns:        int32(config.GetInt("database.min_conns")),
		MaxConnLifetime: time.Minute * time.Duration(config.GetInt("database.max_conn_lifetime_minutes")),
	}
	database.InitializePostgres(pgCfg)
	defer database.ClosePostgres()

	repo := bot.NewRepository(database.GetPostgresDB())

	// Initialize the bot with the database
	botAPI, err := bot.NewBot(botToken, repo, workers)
	if err != nil {
		log.Fatalf("Error initializing bot: %v", err)
	}

	// Start the bot
	log.Println("Bot is starting...")
	ctx := context.Background() // Create a context
	if err := botAPI.Start(ctx); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}
