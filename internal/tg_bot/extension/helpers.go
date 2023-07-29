package extension

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func ErrorMessage(chatID int64) tgbotapi.Chattable {
	return tgbotapi.NewMessage(chatID, "Произошла ошибка, повтори попытку позже")
}
