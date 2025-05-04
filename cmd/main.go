package main

import (
	"database/sql"
	"log"

	"context"
	"qaBot/internal/bot"
	"qaBot/pkg/config"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Retrieve configuration values
	botToken := config.GetString("bot_token")
	dbPath := config.GetString("database.connection_string")

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the bot with the database
	botAPI, err := bot.NewBot(botToken, db)
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
