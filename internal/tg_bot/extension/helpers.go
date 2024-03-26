package extension

import (
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func DefaultErrorMessage(chatID int64) tgbotapi.Chattable {
	return ErrorMessage(chatID, "Произошла ошибка, повторите попытку еще раз")
}

func ErrorMessage(chatID int64, text string) tgbotapi.Chattable {
	return tgbotapi.NewMessage(chatID, text)
}

func CreateChatInviteLink(
	bot *tgbotapi.BotAPI,
	chatID int64,
	nominatorTelegramNickname, NomineeTelegramNickname string,
) (string, error) {
	createInviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
		Name:        fmt.Sprintf("%s -> %s", nominatorTelegramNickname, NomineeTelegramNickname),
		MemberLimit: 1,
	}
	response, err := bot.Request(createInviteLinkConfig)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	if err = json.Unmarshal(response.Result, &data); err != nil {
		return "", err
	}

	if inviteLink, ok := data["invite_link"].(string); ok {
		return inviteLink, nil
	}

	return "", errors.New("could not create invite link")
}
