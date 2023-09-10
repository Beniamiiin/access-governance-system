package commands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	addCommentCommandName = "add_comment"

	waitingForCommentState = "waiting_for_comment"
)

type addCommentCommand struct {
	userRepository     repositories.UserRepository
	proposalRepository repositories.ProposalRepository
	voteBotConfig      configs.VoteBot
	logger             *zap.SugaredLogger
}

func NewAddCommentCommand(
	userRepository repositories.UserRepository,
	proposalRepository repositories.ProposalRepository,
	voteBotConfig configs.VoteBot,
	logger *zap.SugaredLogger,
) Command {
	return &addCommentCommand{
		userRepository:     userRepository,
		proposalRepository: proposalRepository,
		voteBotConfig:      voteBotConfig,
		logger:             logger,
	}
}

func (c *addCommentCommand) CanHandle(command string) bool {
	return command == addCommentCommandName
}

func (c *addCommentCommand) Handle(command string, user *models.User, chatID int64) []tgbotapi.Chattable {
	switch user.TelegramState.LastCommandState {
	case "":
		return []tgbotapi.Chattable{c.handleAddCommentCommand(command, user, chatID)}
	case waitingForCommentState:
		return []tgbotapi.Chattable{c.handleWaitingForCommentState(command, user, chatID)}
	default:
		c.logger.Errorf("user has unknown state: %s", user.TelegramState.LastCommandState)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}
}

func (c *addCommentCommand) handleAddCommentCommand(command string, user *models.User, chatID int64) tgbotapi.Chattable {
	parts := strings.Split(command, ":")
	if len(parts) != 2 {
		c.logger.Errorf("user has invalid command: %s", command)
		return tgbot.DefaultErrorMessage(chatID)
	}

	proposalID, err := strconv.ParseInt(parts[1], 0, 64)
	if err != nil {
		c.logger.Errorf("could not get proposal id: %s", command)
		return tgbot.DefaultErrorMessage(chatID)
	}

	proposal, err := c.proposalRepository.GetOneByID(proposalID)
	if err != nil {
		c.logger.Errorf("could not get proposal: %s", command)
		return tgbot.DefaultErrorMessage(chatID)
	}

	user.TempProposal = *proposal
	user.TelegramState.LastCommandState = waitingForCommentState
	_ = c.updateUser(user)

	return tgbotapi.NewMessage(chatID, "Введите комментарий")
}

func (c *addCommentCommand) handleWaitingForCommentState(comment string, user *models.User, chatID int64) tgbotapi.Chattable {
	bot, err := tgbotapi.NewBotAPI(c.voteBotConfig.Token)
	if err != nil {
		c.logger.Errorf("could not create bot: %v", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	text := fmt.Sprintf("@%s оставил комментарий: %s", user.TelegramNickname, comment)
	message := tgbotapi.NewMessage(int64(user.TempProposal.Poll.ChatID), text)
	message.BaseChat.ReplyToMessageID = user.TempProposal.Poll.PollMessageID

	_, err = bot.Send(message)
	if err != nil {
		c.logger.Errorf("could not send message: %v", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	user.TempProposal = models.Proposal{}
	user.TelegramState = models.TelegramState{}
	_ = c.updateUser(user)

	return tgbotapi.NewMessage(chatID, "Спасибо, ваш комментарий отправлен")
}

func (c *addCommentCommand) updateUser(user *models.User) error {
	_, err := c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
	}
	return err
}
