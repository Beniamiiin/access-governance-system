package extension

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func DefaultErrorMessage(chatID int64) tgbotapi.Chattable {
	return ErrorMessage(chatID, "Произошла ошибка, повторите попытку еще раз")
}

func ErrorMessage(chatID int64, text string) tgbotapi.Chattable {
	return tgbotapi.NewMessage(chatID, text)
}

func CreateChatInviteLink(bot *tgbotapi.BotAPI, chatID int64) (string, error) {
	inviteLinkConfig := tgbotapi.ChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}

	inviteLink, err := bot.GetInviteLink(inviteLinkConfig)
	if err != nil {
		return "", err
	}

	return inviteLink, nil
}
