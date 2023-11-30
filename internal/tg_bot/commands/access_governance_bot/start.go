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
	bot            *tgbotapi.BotAPI
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

func (c *startCommand) Handle(text, arguments string, user *models.User, chatID int64) []tgbotapi.Chattable {
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
		message := c.createInstructionMessageForSeeder(chatID)

		if message == nil {
			return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
		}

		messages = append(messages, message)
	}

	return messages
}

func (c *startCommand) createInstructionMessageForSeeder(chatID int64) tgbotapi.Chattable {
	if c.bot == nil {
		var err error
		c.bot, err = tgbotapi.NewBotAPI(c.config.AccessGovernanceBot.Token)
		if err != nil {
			c.logger.Fatalf("could not create bot: %v", err)
			return nil
		}
	}

	seedersChatInviteLink, err := c.createSeedersChatInviteLink()
	if err != nil {
		c.logger.Fatalf("could not create seeders chat invite link: %v", err)
		return nil
	}

	membersChatInviteLink, err := c.createMembersChatInviteLink()
	if err != nil {
		c.logger.Fatalf("could not create members chat invite link: %v", err)
		return nil
	}

	messageText := fmt.Sprintf(`
Я заметил, что ты являешься сидером, но ты еще не полностью авторизован в нашем сообществе.

Для того, чтобы авторизоваться, зайди в нашу группу для seeders(%s) и для members(%s).
И следуй инструкции, которую ты найдешь в закрепленном сообщении в группе для members.
`, seedersChatInviteLink, membersChatInviteLink)

	message := tgbotapi.NewMessage(chatID, messageText)
	message.DisableWebPagePreview = true

	return message
}

func (c *startCommand) createMembersChatInviteLink() (string, error) {
	inviteLinkConfig := tgbotapi.ChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: c.config.App.MembersChatID,
		},
	}

	inviteLink, err := c.bot.GetInviteLink(inviteLinkConfig)
	if err != nil {
		return "", err
	}

	return inviteLink, nil
}

func (c *startCommand) createSeedersChatInviteLink() (string, error) {
	inviteLinkConfig := tgbotapi.ChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: c.config.App.SeedersChatID,
		},
	}

	inviteLink, err := c.bot.GetInviteLink(inviteLinkConfig)
	if err != nil {
		return "", err
	}

	return inviteLink, nil
}
