package commands

import (
	"access_governance_system/internal"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const pendingProposalsCommandName = "pending_proposals"

type pendingProposalsCommand struct {
	proposalRepository repositories.ProposalRepository
	logger             *zap.SugaredLogger
}

func NewPendingProposalsCommand(proposalRepository repositories.ProposalRepository, logger *zap.SugaredLogger) Command {
	return &pendingProposalsCommand{
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *pendingProposalsCommand) CanHandle(command string) bool {
	return command == pendingProposalsCommandName
}

func (c *pendingProposalsCommand) Handle(text string, user *models.User, chatID int64) []tgbotapi.Chattable {
	proposals, err := c.proposalRepository.GetManyByStatus(models.ProposalStatusCreated)
	if err != nil {
		c.logger.Errorw("failed to get proposals", "error", err)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}

	messages := make([]tgbotapi.Chattable, 0, len(proposals))

	for _, proposal := range proposals {
		var messageText string

		if user.Role == models.UserRoleSeeder {
			messageText += fmt.Sprintf("Тип: %s\n", proposal.NomineeRole.String())
		}

		messageText += fmt.Sprintf("Участник: %s (@%s)\n", proposal.NomineeName, proposal.NomineeTelegramNickname)

		if user.Role == models.UserRoleSeeder {
			messageText += fmt.Sprintf("Комментарий: %s\n", proposal.Comment)
		}

		messageText += fmt.Sprintf("Дата начала: %s\n", internal.Format(proposal.CreatedAt))
		messageText += fmt.Sprintf("Дата окончания: %s\n", internal.Format(proposal.FinishedAt))

		message := tgbotapi.NewMessage(chatID, messageText)

		switch user.Role {
		case models.UserRoleSeeder:
			pollChatID := strings.TrimPrefix(strconv.Itoa(proposal.Poll.ChatID), "-100")

			message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("Проголосовать", fmt.Sprintf("https://t.me/c/%s/%d", pollChatID, proposal.Poll.PollMessageID)),
					tgbotapi.NewInlineKeyboardButtonURL("Обсудить", fmt.Sprintf("https://t.me/c/%s/%d", pollChatID, proposal.Poll.DiscussionMessageID)),
				),
			)
		case models.UserRoleMember:
			message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Оставить комментарий", fmt.Sprintf("add_comment:%d", proposal.ID)),
				),
			)
		}

		messages = append(messages, message)
	}

	if len(messages) == 0 {
		return []tgbotapi.Chattable{tgbotapi.NewMessage(chatID, "Нет предложений на рассмотрении")}
	}

	return messages
}
