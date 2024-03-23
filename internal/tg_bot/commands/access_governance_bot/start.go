package agbcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const startCommandName = "start"

type startCommand struct {
	config         configs.AccessGovernanceBotConfig
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewStartCommand(
	config configs.AccessGovernanceBotConfig,
	userRepository repositories.UserRepository,
	logger *zap.SugaredLogger,
) commands.Command {
	return &startCommand{
		config:         config,
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *startCommand) CanHandle(command string) bool {
	return command == startCommandName
}

func (c *startCommand) Handle(command, arguments string, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	var messages []tgbotapi.Chattable

	text := `
Привет! Я — бот Shmit16 и я помогаю в создании единого чата сообщества Shmit16.

Вот что я умею:
1. /create_proposal — с помощью данной команды, ты можешь создать пригласить нового участника в сообщество.
2. /pending_proposals — с помощью данной команды, ты можешь посмотреть все предложения, которые отправлены на голосование.
`
	messages = append(messages, tgbotapi.NewMessage(chatID, text))

	if user.Role == models.UserRoleSeeder && user.DiscordID == 0 {
		message := c.createInstructionMessageForSeeder(bot, chatID)

		if message == nil {
			return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
		}

		messages = append(messages, message)
	}

	return messages
}

func (c *startCommand) createInstructionMessageForSeeder(bot *tgbotapi.BotAPI, chatID int64) tgbotapi.Chattable {
	seedersChatInviteLink, err := tgbot.CreateChatInviteLink(bot, c.config.App.SeedersChatID)
	if err != nil {
		c.logger.Errorf("could not create seeders chat invite link: %v", err)
		return nil
	}

	membersChatInviteLink, err := tgbot.CreateChatInviteLink(bot, c.config.App.MembersChatID)
	if err != nil {
		c.logger.Errorf("could not create members chat invite link: %v", err)
		return nil
	}

	messageText := fmt.Sprintf(`
Обязательно убедись, что ты вступил в наши группы:
1. Вступить в группу для members — %s
2. Вступить в группу для seeders — %s
`, membersChatInviteLink, seedersChatInviteLink)

	message := tgbotapi.NewMessage(chatID, messageText)
	message.DisableWebPagePreview = true

	return message
}
