package agbhandlers

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"access_governance_system/internal/tg_bot/handlers"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type accessGovernanceBotCommandHandler struct {
	config         configs.AccessGovernanceBotConfig
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger

	commands []commands.Command
}

func NewAccessGovernanceBotCommandHandler(
	config configs.AccessGovernanceBotConfig,
	userRepository repositories.UserRepository,
	logger *zap.SugaredLogger,
	commands []commands.Command,
) handlers.CommandHandler {
	return &accessGovernanceBotCommandHandler{
		config:         config,
		userRepository: userRepository,
		logger:         logger,
		commands:       commands,
	}
}

func (h *accessGovernanceBotCommandHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) []tgbotapi.Chattable {
	h.logger.Info("received message")

	message := update.Message
	callbackQuery := update.CallbackQuery

	if message == nil && callbackQuery == nil {
		h.logger.Warn("received unknown updates")
		return []tgbotapi.Chattable{}
	}

	var (
		chatID       int64
		telegramUser *tgbotapi.User
	)

	if message != nil {
		chatID = message.Chat.ID
		telegramUser = message.From
	} else if callbackQuery != nil {
		chatID = callbackQuery.Message.Chat.ID
		telegramUser = callbackQuery.From
	}

	if len(message.NewChatMembers) > 0 {
		return h.handleNewChatMembers(bot, message)
	}

	if telegramUser.ID != chatID {
		h.logger.Infow("received message", "message", message)
		return []tgbotapi.Chattable{}
	}

	user, errMessage := h.createUserIfNeeded(telegramUser, chatID)
	if errMessage != nil {
		return []tgbotapi.Chattable{errMessage}
	}

	if message != nil {
		h.logger.Infow("received message", "message", message)
		if message.IsCommand() {
			h.logger.Infow("received command", "command", message.Command())
			return h.tryToHandleCommand(message.Command(), h.commands, user, bot, chatID)
		} else if user.TelegramState.LastCommand != "" {
			h.logger.Infow("received subcommand", "subcommand", message.Text)
			return h.tryToHandleSubCommand(user.TelegramState.LastCommand, message.Text, h.commands, user, bot, chatID)
		}
	}

	if callbackQuery != nil {
		h.logger.Infow("received callback query", "callback_query", callbackQuery)
		return h.tryToHandleQueryCallback(callbackQuery.Data, h.commands, user, bot, chatID)
	}

	h.logger.Warn("received unknown message")
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) createUserIfNeeded(telegramUser *tgbotapi.User, chatID int64) (*models.User, tgbotapi.Chattable) {
	user, err := h.userRepository.GetOneByTelegramID(telegramUser.ID)
	if err != nil {
		h.logger.Warnw("failed to get user", "error", err)
	}

	if user == nil {
		isItSeeder := false

		for _, seederName := range h.config.App.InitialSeeders {
			if telegramUser.UserName == seederName {
				user = &models.User{
					Name: func() string {
						var parts []string

						if telegramUser.FirstName != "" {
							parts = append(parts, telegramUser.FirstName)
						}

						if telegramUser.LastName != "" {
							parts = append(parts, telegramUser.LastName)
						}

						return strings.Join(parts, " ")
					}(),
					TelegramID:       telegramUser.ID,
					TelegramNickname: telegramUser.UserName,
					Role:             models.UserRoleSeeder,
				}

				user, err = h.userRepository.Create(user)
				if err != nil {
					h.logger.Errorw("failed to create user", "error", err)
					return nil, tgbot.DefaultErrorMessage(chatID)
				}

				isItSeeder = true
				break
			}
		}

		if !isItSeeder {
			return nil, tgbotapi.NewMessage(chatID, fmt.Sprintf("Привет! К сожалению, ты не участник сообщества %s.", h.config.App.CommunityName))
		}
	}

	return user, nil
}

func (h *accessGovernanceBotCommandHandler) tryToHandleCommand(command string, commands []commands.Command, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	for _, handler := range commands {
		if handler.CanHandle(command) {
			user.TempProposal = models.Proposal{}
			user.TelegramState = models.TelegramState{LastCommand: command}

			user, err := h.userRepository.Update(user)
			if err != nil {
				h.logger.Errorw("failed to update user", "error", err)
			}

			return handler.Handle(command, "", user, bot, chatID)
		}
	}

	h.logger.Warnw("received unknown command", "command", command)
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) tryToHandleSubCommand(command, subCommand string, commands []commands.Command, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	command = strings.Split(command, ":")[0]

	for _, handler := range commands {
		if handler.CanHandle(command) {
			responseMessage := handler.Handle(subCommand, "", user, bot, chatID)
			if responseMessage == nil {
				h.logger.Errorw("failed to handle subcommand", "subCommand", subCommand)
				break
			}

			return responseMessage
		}
	}

	h.logger.Errorf("received unknown subcommand: %s for command: %s", subCommand, command)
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) tryToHandleQueryCallback(query string, commands []commands.Command, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	parts := strings.Split(query, ":")
	if len(parts) == 0 {
		h.logger.Error("received empty query callback")
		return []tgbotapi.Chattable{}
	}

	command := parts[0]

	for _, handler := range commands {
		if handler.CanHandle(command) {
			user.TelegramState.LastCommand = query

			user, err := h.userRepository.Update(user)
			if err != nil {
				h.logger.Errorw("failed to update user", "error", err)
			}

			return handler.Handle(query, "", user, bot, chatID)
		}
	}

	h.logger.Errorw("received unknown command", "command", command)
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) handleNewChatMembers(bot *tgbotapi.BotAPI, message *tgbotapi.Message) []tgbotapi.Chattable {
	messages := []tgbotapi.Chattable{}

	for _, newChatMember := range message.NewChatMembers {
		user, err := h.userRepository.GetOneByTelegramNickname(newChatMember.UserName)
		if user == nil || err != nil {
			h.logger.Errorw("failed to get user", "error", err)
			continue
		}

		user.TelegramID = newChatMember.ID

		_, err = h.userRepository.Update(user)
		if err != nil {
			h.logger.Errorw("failed to update user", "error", err)
			continue
		}

		var messageText string

		if user.Role == models.UserRoleSeeder {
			seedersChatInviteLink, err := tgbot.CreateChatInviteLink(bot, h.config.App.SeedersChatID)
			if err != nil {
				h.logger.Fatalf("could not create seeders chat invite link: %v", err)
				continue
			}

			messageText = fmt.Sprintf(`
Привет, %s! Добро пожаловать в сообщество %s.

Для того, чтобы авторизоваться тебе надо:
1. Вступить в нашу группу для seeders - %s
2. Подключиться к нашему discord серверу - %s
3. Отправить команду %s в чате в discord
`, newChatMember.FirstName, h.config.App.CommunityName, seedersChatInviteLink, h.config.DiscordInviteLink, "`!authorize`")
		} else if user.Role == models.UserRoleMember {
			messageText = fmt.Sprintf(`
Привет, %s! Добро пожаловать в сообщество %s.

Для того, чтобы авторизоваться тебе надо:
1. Подключиться к нашему discord серверу - %s
2. Отправить команду %s в чате в discord
`, newChatMember.FirstName, h.config.App.CommunityName, h.config.DiscordInviteLink, "`!authorize`")
		}

		message := tgbotapi.NewMessage(newChatMember.ID, messageText)
		message.DisableWebPagePreview = true
		message.ParseMode = tgbotapi.ModeMarkdown
		messages = append(messages, message)
	}

	return messages
}
