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
	questions  []Question
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

// InitQuestions initializes the bot's questions from the database.
func (b *Bot) InitQuestions() error {
	if b.repository == nil {
		return errors.New("repository is not initialized")
	}

	questions, err := b.repository.GetQuestions()
	if err != nil {
		return err
	}

	b.questions = questions
	return nil
}

// Start begins listening for updates and initializes questions from the database.
func (b *Bot) Start(ctx context.Context) error {
	if b.api == nil {
		return ErrBotNotInitialized
	}

	// Initialize questions from the database
	if err := b.InitQuestions(); err != nil {
		return err
	}

	log.Println("Questions initialized from the database.")

	// Set up a handler for the /getquestions command
	b.api.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/getquestions",
		tgbot.MatchTypeExact,
		b.GetQuestions,
	)

	log.Println("Bot started")
	// Start polling for updates
	b.api.Start(ctx)
	return nil
}
