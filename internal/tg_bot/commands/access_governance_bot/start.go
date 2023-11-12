package agbcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const startCommandName = "start"

type startCommand struct {
	appConfig      configs.App
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewStartCommand(appConfig configs.App, userRepository repositories.UserRepository, logger *zap.SugaredLogger) commands.Command {
	return &startCommand{
		appConfig:      appConfig,
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *startCommand) CanHandle(command string) bool {
	return command == startCommandName
}

func (c *startCommand) Handle(text, arguments string, user *models.User, chatID int64) []tgbotapi.Chattable {
	parseMode := tgbotapi.ModeMarkdownV2

	messageText := tgbotapi.EscapeText(parseMode, fmt.Sprintf(`
Привет! Я - бот %s, и вот что я могу предложить:

/create_proposal - с помощью данной команды, ты можешь создать новое предложение о добавлении нового участника в сообщество.
/pending_proposals - с помощью данной команды, ты можешь посмотреть все предложения, который находятся в данный момент на голосовании.
/approved_proposals - с помощью данной команды, ты можешь посмотреть все принятые/отклоненные предложения.

Для более подробного изучения всех доступных команд, нажми на кнопку Menu.
`, c.appConfig.CommunityName))
	messageText = strings.Replace(messageText, "Menu", "*Menu*", -1)
	message := tgbotapi.NewMessage(chatID, messageText)
	message.ParseMode = parseMode
	return []tgbotapi.Chattable{message}
}
