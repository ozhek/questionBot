package bot

import (
	"context"
	"fmt"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
)

type Question struct {
	ID           int        `json:"id"`
	Text         string     `json:"text"`
	Answer       string     `json:"answer"`
	SubQuestions []Question `json:"sub_questions,omitempty"`
}

func (b *Bot) GetQuestions(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	replyKeyboard := reply.New(
		reply.WithPrefix("questions_keyboard"),
		reply.IsSelective(),
		reply.IsOneTimeKeyboard(),
	)

	for i, question := range b.questions {
		replyKeyboard = replyKeyboard.
			Button(fmt.Sprintf("%d. %s", i+1, question.Text), tbot, tgbot.MatchTypeExact, b.onReplyKeyboardSelect).
			Row()
	}

	tbot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Select question:",
		ReplyMarkup: replyKeyboard,
	})
}

func (b *Bot) onReplyKeyboardSelect(ctx context.Context, tbot *tgbot.Bot, update *models.Update) {
	var (
		pref string
		q    Question
	)

	if update.Message.Text != "" {
		fmt.Sscanf(update.Message.Text, "%s %s", &pref)
	}

	q, err := b.getQuestionByIndexes(pref)
	if err != nil {
		tbot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "I don't understand the question",
		})
	}

	smp := tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   q.Answer,
	}

	replyKeyboard := reply.New(
		reply.WithPrefix("subquestions_keyboard"),
		reply.IsSelective(),
		reply.IsOneTimeKeyboard(),
	)

	for i, question := range q.SubQuestions {
		replyKeyboard = replyKeyboard.
			Button(fmt.Sprintf("%s%d. %s", pref, i+1, question.Text), tbot, tgbot.MatchTypeExact, b.onReplyKeyboardSelect).
			Row()
	}

	if len(q.SubQuestions) != 0 {
		smp.ReplyMarkup = replyKeyboard
		smp.Text += "\n**You can also choose subquestions**"
	}

	tbot.SendMessage(ctx, &smp)
}

func (b *Bot) getQuestionByIndexes(s string) (Question, error) {
	if s == "" {
		return Question{}, ErrQuestionNotFound
	}

	var (
		idx  int
		qs   = b.questions
		idxs = strings.Split(s, ".")
	)

	for i, part := range idxs {
		fmt.Sscanf(part, "%d", &idx)
		idx--

		if idx < 0 || idx >= len(qs) {
			return Question{}, ErrQuestionNotFound
		}

		if i == len(idxs)-2 {
			break
		}

		qs = qs[idx].SubQuestions
	}

	return qs[idx], nil
}
