package extension

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func DefaultErrorMessage(chatID int64) tgbotapi.Chattable {
	return ErrorMessage(chatID, "Произошла ошибка, повторите попытку еще раз")
}

func ErrorMessage(chatID int64, text string) tgbotapi.Chattable {
	return tgbotapi.NewMessage(chatID, text)
}
