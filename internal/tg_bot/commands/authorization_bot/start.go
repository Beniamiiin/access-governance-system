package abcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	"access_governance_system/internal/tg_bot/extension"
	"strconv"

	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const startCommandName = "start"

type startCommand struct {
	discord        *discordgo.Session
	config         configs.Discord
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewStartCommand(config configs.Discord, userRepository repositories.UserRepository, logger *zap.SugaredLogger) commands.Command {
	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		logger.Fatalw("failed to create discord session", "error", err)
	}

	return &startCommand{
		discord:        discord,
		config:         config,
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *startCommand) CanHandle(command string) bool {
	return command == startCommandName
}

func (c *startCommand) Handle(text, discordID string, user *models.User, chatID int64) []tgbotapi.Chattable {
	if user.Role == models.UserRoleMember || user.Role == models.UserRoleSeeder {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chatID, "Привет, ты уже авторизован."),
		}
	}

	err := c.discord.GuildMemberRoleAdd(c.config.ChannelID, discordID, c.config.MemberRoleID)
	if err != nil {
		c.logger.Errorw("failed to add a role", "error", err)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	}

	user.DiscordID, err = strconv.Atoi(discordID)
	if err != nil {
		c.logger.Errorw("failed to parse discord id", "error", err)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	}

	_, err = c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	}

	err = c.discord.GuildMemberRoleRemove(c.config.ChannelID, discordID, c.config.GuestRoleID)
	if err != nil {
		c.logger.Warnw("failed to remove a role", "error", err)
	}

	message := tgbotapi.NewMessage(chatID, "Привет, ты успешно авторизован, можешь возвращаться в Discord")
	return []tgbotapi.Chattable{message}
}
