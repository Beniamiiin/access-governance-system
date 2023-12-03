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

func (c *startCommand) Handle(command, discordID string, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	if (user.Role == models.UserRoleMember || user.Role == models.UserRoleSeeder) && user.DiscordID != 0 {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chatID, "Привет, ты успешно авторизован, можешь возвращаться в Discord"),
		}
	}

	err := c.discord.GuildMemberRoleAdd(c.config.ServerID, discordID, c.config.MemberRoleID)
	if err != nil {
		c.logger.Errorw(
			"failed to add a role",
			"server_id", c.config.ServerID,
			"discord_id", discordID,
			"role_id", c.config.MemberRoleID,
			"error", err,
		)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	}

	user.DiscordID, err = strconv.Atoi(discordID)
	if err != nil {
		c.logger.Errorw("failed to parse discord id", "discord_id", discordID, "error", err)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	}

	if user.Role == models.UserRoleGuest {
		user.Role = models.UserRoleMember
	}

	_, err = c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "user", user, "error", err)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	}

	message := tgbotapi.NewMessage(chatID, "Привет, ты успешно авторизован, можешь возвращаться в Discord")
	return []tgbotapi.Chattable{message}
}
