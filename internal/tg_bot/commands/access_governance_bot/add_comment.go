package agbcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
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
	voteBotConfig      configs.Bot
	logger             *zap.SugaredLogger
}

func NewAddCommentCommand(
	userRepository repositories.UserRepository,
	proposalRepository repositories.ProposalRepository,
	voteBotConfig configs.Bot,
	logger *zap.SugaredLogger,
) commands.Command {
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

func (c *addCommentCommand) Handle(command, arguments string, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	switch user.TelegramState.LastCommandState {
	case "":
		return []tgbotapi.Chattable{c.handleAddCommentCommand(command, user, chatID)}
	case waitingForCommentState:
		return []tgbotapi.Chattable{c.handleWaitingForCommentState(command, user, chatID)}
	default:
		c.logger.Errorw("user has unknown state", "state", user.TelegramState.LastCommandState)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}
}

func (c *addCommentCommand) handleAddCommentCommand(command string, user *models.User, chatID int64) tgbotapi.Chattable {
	parts := strings.Split(command, ":")
	if len(parts) != 2 {
		c.logger.Errorw("user has invalid command", "command", command)
		return tgbot.DefaultErrorMessage(chatID)
	}

	proposalID, err := strconv.ParseInt(parts[1], 0, 64)
	if err != nil {
		c.logger.Errorw("could not get proposal id", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	proposal, err := c.proposalRepository.GetOneByID(proposalID)
	if err != nil {
		c.logger.Errorw("could not get proposal", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	user.TempProposal = *proposal
	user.TelegramState.LastCommandState = waitingForCommentState
	_ = c.updateUser(user)

	text := fmt.Sprintf("Введи комментарий, почему ты считаешь, что %s (@%s) стоит добавить. Чем подробнее, тем лучше мы сможем понять твою точку зрения.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	return tgbotapi.NewMessage(chatID, text)
}

func (c *addCommentCommand) handleWaitingForCommentState(comment string, user *models.User, chatID int64) tgbotapi.Chattable {
	bot, err := tgbotapi.NewBotAPI(c.voteBotConfig.Token)
	if err != nil {
		c.logger.Errorw("could not create bot", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	text := fmt.Sprintf("@%s оставил комментарий: %s", user.TelegramNickname, comment)
	message := tgbotapi.NewMessage(int64(user.TempProposal.Poll.ChatID), text)
	message.BaseChat.ReplyToMessageID = user.TempProposal.Poll.PollMessageID

	_, err = bot.Send(message)
	if err != nil {
		c.logger.Errorw("could not send message", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	user.TempProposal = models.Proposal{}
	user.TelegramState = models.TelegramState{}
	_ = c.updateUser(user)

	return tgbotapi.NewMessage(chatID, "Спасибо, твой комментарий добавлен к заявке.")
}

func (c *addCommentCommand) updateUser(user *models.User) error {
	_, err := c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
	}
	return err
}
