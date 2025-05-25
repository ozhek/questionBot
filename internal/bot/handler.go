package bot

import (
	"context"
	"database/sql"
	"errors"
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

type PendingQuestionData struct {
	ParentID int
	Lang     string
	EditID   *int // nil if adding
}

func (b *Bot) GetStart(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	commands := []string{
		"/start - Show this help message",
		"/questions - List available questions",
		"/language - Set language",
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

	fmt.Printf("GetQuestions called by user %d\n", update.Message.From.ID)

	isAdmin := adminIDs[update.Message.From.ID]

	questions, err := b.getQuestionsByUserID(ctx, update.Message.From.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No questions available.",
		})
		return
	}

	keyboard := b.buildQuestionKeyboard(questions, 0, 0, pageSize, isAdmin)

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

func (b *Bot) buildQuestionKeyboard(questions []Question, parentID, page, pageSize int, isAdmin bool) *models.InlineKeyboardMarkup {
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
		if isAdmin {
			rows = append(rows,
				[]models.InlineKeyboardButton{
					{
						Text:         "✏️",
						CallbackData: fmt.Sprintf("edit_%d", q.ID),
					},
					{
						Text:         "🗑️",
						CallbackData: fmt.Sprintf("del_%d", q.ID),
					},
				})
		}
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

	// Add "Add Question" button for admins at the root level
	if isAdmin {
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         "➕ Add Question",
				CallbackData: fmt.Sprintf("add_question_%d", parentID),
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

	fmt.Printf("HandleQuestionCallback received: %s from user %d\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)

	isAdmin := adminIDs[update.CallbackQuery.From.ID]

	data := update.CallbackQuery.Data

	id, err := strconv.Atoi(strings.TrimPrefix(data, "q_"))
	if err != nil {
		return
	}

	q, err := b.repository.GetQuestionByID(id)
	if err != nil {
		return
	}

	keyboard := b.buildQuestionKeyboard(q.SubQuestions, q.ID, 0, pageSize, isAdmin)

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

	fmt.Printf("HandleQuestionPageCallback received: %s from user %d\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)

	userID := update.CallbackQuery.From.ID
	isAdmin := adminIDs[userID]

	data := update.CallbackQuery.Data

	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return
	}
	parentID, _ := strconv.Atoi(parts[1])
	page, _ := strconv.Atoi(parts[2])

	questions, err := b.getQuestionsByUserID(ctx, userID)
	if err != nil {
		return
	}

	if parentID != 0 {
		parentQ, err := b.repository.GetQuestionByID(parentID)
		if err != nil {
			return
		}
		questions = parentQ.SubQuestions
	}

	keyboard := b.buildQuestionKeyboard(questions, parentID, page, pageSize, isAdmin)

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

	fmt.Printf("HandleQuestionBackCallback received: %s from user %d\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)

	isAdmin := adminIDs[update.CallbackQuery.From.ID]
	data := update.CallbackQuery.Data

	childID, _ := strconv.Atoi(strings.TrimPrefix(data, "back_"))
	if childID == 0 {
		return
	}

	currentQ, err := b.repository.GetQuestionByID(childID)
	if err != nil {
		return
	}

	if currentQ.ParentID == 0 {
		questions, err := b.getQuestionsByUserID(ctx, update.CallbackQuery.From.ID)
		if err != nil {
			return
		}
		keyboard := b.buildQuestionKeyboard(questions, 0, 0, pageSize, isAdmin)
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

	keyboard := b.buildQuestionKeyboard(parentQ.SubQuestions, parentQ.ID, 0, pageSize, isAdmin)

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

func (b *Bot) HandleAddQuestion(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	fmt.Printf("HandleAddQuestion received: %s from user %d\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)

	userID := update.CallbackQuery.From.ID
	if !adminIDs[userID] {
		return
	}

	data := update.CallbackQuery.Data
	parentID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_question_"))

	lang, err := b.repository.GetUserLang(userID)
	if err != nil || lang == "" {
		lang = "en"
	}

	b.pendingMutex.Lock()
	b.pendingQuestionEdits[userID] = &PendingQuestionData{
		ParentID: parentID,
		Lang:     lang,
		EditID:   nil,
	}
	b.pendingMutex.Unlock()

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   fmt.Sprintf("Send your new question for language [%s] and parent [%d] in the format:\n\nquestion|answer", lang, parentID),
	})
}

func (b *Bot) HandleEditQuestion(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	fmt.Printf("HandleEditQuestion received: %s from user %d\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)

	userID := update.CallbackQuery.From.ID
	if !adminIDs[userID] {
		return
	}

	data := update.CallbackQuery.Data
	id, err := strconv.Atoi(strings.TrimPrefix(data, "edit_"))
	if err != nil {
		return
	}

	b.pendingMutex.Lock()
	b.pendingQuestionEdits[userID] = &PendingQuestionData{
		EditID: &id,
	}
	b.pendingMutex.Unlock()

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   fmt.Sprintf("Send edited text for question #%d in format:\n\nquestion|answer", id),
	})
}

func (b *Bot) HandleDeleteQuestion(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	fmt.Printf("HandleDeleteQuestion received: %s from user %d\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)

	userID := update.CallbackQuery.From.ID
	if !adminIDs[userID] {
		return
	}

	data := update.CallbackQuery.Data
	id, err := strconv.Atoi(strings.TrimPrefix(data, "del_"))
	if err != nil {
		return
	}

	err = b.repository.DeleteQuestionByID(id)
	if err != nil {
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "Failed to delete question.",
		})
		return
	}

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   fmt.Sprintf("Question #%d deleted.", id),
	})
}

// HandleMessageInput processes text input for adding or editing questions.
func (b *Bot) HandleMessageInput(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	fmt.Printf("HandleMessageInput received from user %d: %s\n", update.Message.From.ID, update.Message.Text)

	userID := update.Message.From.ID

	b.pendingMutex.RLock()
	session, ok := b.pendingQuestionEdits[userID]
	b.pendingMutex.RUnlock()
	if !ok || update.Message.Text == "" {
		return
	}

	parts := strings.SplitN(update.Message.Text, "|", 2)
	if len(parts) != 2 {
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Invalid format. Use:\n\nquestion|answer",
		})
		return
	}
	questionText := strings.TrimSpace(parts[0])
	answerText := strings.TrimSpace(parts[1])

	if session.EditID != nil {
		// Update existing question
		err := b.repository.UpdateQuestion(*session.EditID, questionText, answerText)
		if err != nil {
			tbot.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Failed to update question.",
			})
			return
		}
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Question updated successfully.",
		})
	} else {
		// Create new question
		err := b.repository.CreateQuestion(session.Lang, questionText, answerText, session.ParentID)
		if err != nil {
			tbot.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Failed to create question.",
			})
			return
		}
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Question created successfully.",
		})
	}

	// Clear session
	b.pendingMutex.Lock()
	delete(b.pendingQuestionEdits, userID)
	b.pendingMutex.Unlock()
}
