package agbcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const startCommandName = "start"

type startCommand struct {
	config         configs.AccessGovernanceBotConfig
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewStartCommand(config configs.AccessGovernanceBotConfig, userRepository repositories.UserRepository, logger *zap.SugaredLogger) commands.Command {
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
	var messages = []tgbotapi.Chattable{}

	parseMode := tgbotapi.ModeMarkdownV2

	messageText := tgbotapi.EscapeText(parseMode, fmt.Sprintf(`
Привет! Я - бот %s, и вот что я могу предложить:

/create_proposal - с помощью данной команды, ты можешь создать новое предложение о добавлении нового участника в сообщество.
/pending_proposals - с помощью данной команды, ты можешь посмотреть все предложения, который находятся в данный момент на голосовании.
/approved_proposals - с помощью данной команды, ты можешь посмотреть все принятые/отклоненные предложения.

Для более подробного изучения всех доступных команд, нажми на кнопку Menu.
`, c.config.App.CommunityName))
	messageText = strings.Replace(messageText, "Menu", "*Menu*", -1)
	message := tgbotapi.NewMessage(chatID, messageText)
	message.ParseMode = parseMode
	messages = append(messages, message)

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
Я заметил, что ты являешься сидером, но ты еще не полностью авторизован в нашем сообществе.

Для того, чтобы авторизоваться тебе надо:
1. Вступить в нашу группу для seeders - %s
2. Вступить в нашу группу для members - %s
3. Подключиться к нашему discord серверу - %s
4. Отправить команду %s в чате в discord
`, seedersChatInviteLink, membersChatInviteLink, c.config.DiscordInviteLink, "`!authorize`")

	message := tgbotapi.NewMessage(chatID, messageText)
	message.DisableWebPagePreview = true
	message.ParseMode = tgbotapi.ModeMarkdown

	return message
}
