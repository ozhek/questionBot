package bot

import (
	"context"
	"database/sql"
	"errors"
	"log"

	tgbot "github.com/go-telegram/bot"
)

// Errors
var (
	ErrEmptyToken        = errors.New("bot token cannot be empty")
	ErrBotNotInitialized = errors.New("bot is not initialized")
	ErrQuestionNotFound  = errors.New("question not found")
)

type Bot struct {
	api        *tgbot.Bot
	repository *Repository
}

// NewBot initializes a new Bot instance with the provided token and database.
func NewBot(token string, db *sql.DB) (*Bot, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	bot, err := tgbot.New(token)
	if err != nil {
		return nil, err
	}

	repo := NewRepository(db)

	return &Bot{api: bot, repository: repo}, nil
}

// Start begins listening for updates and initializes questions from the database.
func (b *Bot) Start(ctx context.Context) error {
	if b.api == nil {
		return ErrBotNotInitialized
	}

	// Set up a handler for the /questions command
	b.api.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/questions",
		tgbot.MatchTypeExact,
		b.GetQuestions,
	)

	b.api.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		b.GetStart,
	)

	b.api.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/language",
		tgbot.MatchTypeExact,
		b.HandleLanguage,
	)

	b.api.RegisterHandler(
		tgbot.HandlerTypeCallbackQueryData,
		"q_",
		tgbot.MatchTypePrefix,
		b.HandleQuestionCallback,
	)

	b.api.RegisterHandler(
		tgbot.HandlerTypeCallbackQueryData,
		"p_",
		tgbot.MatchTypePrefix,
		b.HandleQuestionPageCallback,
	)

	b.api.RegisterHandler(
		tgbot.HandlerTypeCallbackQueryData,
		"back_",
		tgbot.MatchTypePrefix,
		b.HandleQuestionBackCallback,
	)

	log.Println("Bot is starting")
	b.api.Start(ctx)
	return nil
}
