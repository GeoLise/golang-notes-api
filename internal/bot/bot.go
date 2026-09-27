package bot

import (
	"context"
	"log"
	"notes-api/internal/storage"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Run(ctx context.Context, token string, logincodesStorage *storage.LoginCodes, usersStorage *storage.Users) error {

	bot, err := tgbot.New(token, tgbot.WithDefaultHandler(defaultHandler))
	if err != nil {
		panic(err)
	}

	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "start", tgbot.MatchTypeCommandStartOnly, startHandler(logincodesStorage, usersStorage))

	bot.Start(ctx)

	return nil
}

func startHandler(logincodesStorage *storage.LoginCodes, usersStorage *storage.Users) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		msg := update.Message
		if msg == nil || msg.From == nil {
			return
		}

		reply := func(text string) {
			b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: msg.Chat.ID, Text: text})
		}

		code := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/start"))
		if code == "" {
			reply("Откройте бота через кнопку «Войти» в приложении.")
			return
		}

		pending, err := logincodesStorage.IsPending(ctx, code)
		if err != nil || !pending {
			reply("Ссылка устарела. Нажмите «Войти» в приложении ещё раз.")
			return
		}

		user, err := usersStorage.UpsertByTelegram(ctx, msg.From.ID, msg.From.Username, msg.From.FirstName)
		if err != nil {
			log.Print("upsert telegram user: ", err)
			reply("Что-то пошло не так, попробуйте позже.")
			return
		}

		if err := logincodesStorage.Confirm(ctx, code, user.Id); err != nil {
			log.Print("confirm login code: ", err)
			reply("Что-то пошло не так, попробуйте позже.")
			return
		}

		reply("Готово! Вернитесь в приложение.")
	}
}

func defaultHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Вот это да! Че реально?",
	})
}
