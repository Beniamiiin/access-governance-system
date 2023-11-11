package abcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	"fmt"

	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const startCommandName = "start"

type startCommand struct {
	discord        *discordgo.Session
	discordConfig  configs.Discord
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewStartCommand(discordConfig configs.Discord, userRepository repositories.UserRepository, logger *zap.SugaredLogger) commands.Command {
	discord, err := discordgo.New("Bot " + discordConfig.Token)
	if err != nil {
		logger.Fatalw("failed to create discord session", "error", err)
	}

	return &startCommand{
		discord:        discord,
		discordConfig:  discordConfig,
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *startCommand) CanHandle(command string) bool {
	return command == startCommandName
}

func (c *startCommand) Handle(text string, user *models.User, chatID int64) []tgbotapi.Chattable {
	if user.Role == models.UserRoleMember || user.Role == models.UserRoleSeeder {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chatID, "Привет, ты уже авторизован."),
		}
	}

	c.discord.ChannelMessageSend(c.discordConfig.AuthorizationChannelID, fmt.Sprintf("Пользователь с Telegram ID %d has started authorization.", user.TelegramID))

	message := tgbotapi.NewMessage(chatID, "Привет, cпасибо, ты авторизован, можешь возвращаться в Discord")
	return []tgbotapi.Chattable{message}
}
