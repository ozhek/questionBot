package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
)

var (
	adminIDs = map[int64]bool{
		687353891: true,
	}
)

const (
	pageSize = 5
)

type Question struct {
	ID           int        `json:"id"`
	Lang         string     `json:"lang"`
	Text         string     `json:"text"`
	Answer       string     `json:"answer"`
	ParentID     int        `json:"parent_id"`
	SubQuestions []Question `json:"sub_questions,omitempty"`
}

func (b *Bot) GetStart(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	commands := []string{
		"/start - Show this help message",
		"/questions - List available questions",
	}

	if adminIDs[userID] {
		commands = append(commands,
			"/createquestion - Create a new question",
			"/addsubquestion - Add a subquestion to a question",
		)
	}

	msg := "Available commands:\n" + strings.Join(commands, "\n")

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})
}

func (b *Bot) GetQuestions(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	questions, err := b.getQuestionsByUserID(ctx, update.Message.From.ID)
	if err != nil || len(questions) == 0 {
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No questions available.",
		})
		return
	}

	keyboard := b.buildQuestionKeyboard(questions, 0, 0, pageSize)

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Choose a question:",
		ReplyMarkup: keyboard,
	})
}

func (b *Bot) getQuestionsByUserID(ctx context.Context, userID int64) ([]Question, error) {
	lang, err := b.repository.GetUserLang(userID)
	if err != nil || lang == "" {
		lang = "en"
	}

	questions, err := b.repository.GetQuestionsByLang(lang)
	if err != nil {
		return []Question{}, err
	}
	return questions, nil

}

func (b *Bot) buildQuestionKeyboard(questions []Question, parentID, page, pageSize int) *models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	// Фильтруем по родителю
	var filtered []Question
	for _, q := range questions {
		if q.ParentID == parentID {
			filtered = append(filtered, q)
		}
	}

	total := len(filtered)
	start := page * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	pageQuestions := filtered[start:end]

	for _, q := range pageQuestions {
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         q.Text,
				CallbackData: fmt.Sprintf("q_%d", q.ID),
			},
		})
	}

	// Пагинация
	var navRow []models.InlineKeyboardButton
	if page > 0 {
		navRow = append(navRow, models.InlineKeyboardButton{
			Text:         "⬅️ Prev",
			CallbackData: fmt.Sprintf("p_%d_%d", parentID, page-1),
		})
	}
	if end < total {
		navRow = append(navRow, models.InlineKeyboardButton{
			Text:         "➡️ Next",
			CallbackData: fmt.Sprintf("p_%d_%d", parentID, page+1),
		})
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	// Назад
	if parentID != 0 {
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         "🔙 Back",
				CallbackData: fmt.Sprintf("back_%d", parentID),
			},
		})
	}

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (b *Bot) HandleQuestionCallback(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data

	id, err := strconv.Atoi(strings.TrimPrefix(data, "q_"))
	if err != nil {
		return
	}

	q, err := b.repository.GetQuestionByID(id)
	if err != nil {
		return
	}

	keyboard := b.buildQuestionKeyboard(q.SubQuestions, q.ID, 0, 5)

	tbot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        fmt.Sprintf("*%s*\n\n%s", q.Text, q.Answer),
		ParseMode:   "Markdown",
		ReplyMarkup: keyboard,
	})
}

func (b *Bot) HandleQuestionPageCallback(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	data := update.CallbackQuery.Data

	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return
	}
	parentID, _ := strconv.Atoi(parts[1])
	page, _ := strconv.Atoi(parts[2])

	parentQ, err := b.repository.GetQuestionByID(parentID)
	if err != nil {
		return
	}

	keyboard := b.buildQuestionKeyboard(parentQ.SubQuestions, parentID, page, 5)

	tbot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: keyboard,
	})
}

func (b *Bot) HandleQuestionBackCallback(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	data := update.CallbackQuery.Data

	fmt.Println("-------------")
	fmt.Println(data)
	fmt.Println("-------------")

	childID, _ := strconv.Atoi(strings.TrimPrefix(data, "back_"))
	if childID == 0 {
		return
	}

	currentQ, err := b.repository.GetQuestionByID(childID)
	if err != nil {
		return
	}

	fmt.Println("-------current------")
	fmt.Println(currentQ)
	fmt.Println("-------------")

	if currentQ.ParentID == 0 {
		questions, err := b.getQuestionsByUserID(ctx, update.CallbackQuery.From.ID)
		if err != nil {
			return
		}
		keyboard := b.buildQuestionKeyboard(questions, 0, 0, pageSize)
		tbot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        "Choose a question:",
			ParseMode:   "Markdown",
			ReplyMarkup: keyboard,
		})
		return
	}

	parentQ, err := b.repository.GetQuestionByID(currentQ.ParentID)
	if err != nil {
		return
	}

	fmt.Println("-------parent------")
	fmt.Println(parentQ)
	fmt.Println("-------------")

	keyboard := b.buildQuestionKeyboard(parentQ.SubQuestions, parentQ.ID, 0, pageSize)

	tbot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "Choose a question:",
		ParseMode:   "Markdown",
		ReplyMarkup: keyboard,
	})
}

func (b *Bot) HandleLanguage(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	replyKeyboard := reply.New(
		reply.WithPrefix("questions_keyboard"),
		reply.IsSelective(),
		reply.IsOneTimeKeyboard(),
	)

	replyKeyboard = replyKeyboard.
		Button("English", tbot, tgbot.MatchTypeExact, b.HandleLanguageSelection).
		Button("Русский", tbot, tgbot.MatchTypeExact, b.HandleLanguageSelection)

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Please choose your language / Пожалуйста, выберите язык:",
		ReplyMarkup: replyKeyboard,
	})
}

func (b *Bot) HandleLanguageSelection(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	var lang string
	switch update.Message.Text {
	case "English":
		lang = "en"
	case "Русский":
		lang = "ru"
	default:
		return
	}
	if err := b.repository.SetUserLang(userID, lang); err == nil {
		msg := map[string]string{
			"en": "Language set to English.",
			"ru": "Язык установлен на русский.",
		}
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   msg[lang],
		})
	}
}
