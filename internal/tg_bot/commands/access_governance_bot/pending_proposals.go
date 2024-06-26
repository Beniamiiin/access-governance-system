package agbcommands

import (
	"access_governance_system/internal"
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

const pendingProposalsCommandName = "pending_proposals"

type pendingProposalsCommand struct {
	userRepository     repositories.UserRepository
	proposalRepository repositories.ProposalRepository
	logger             *zap.SugaredLogger
}

func NewPendingProposalsCommand(
	userRepository repositories.UserRepository,
	proposalRepository repositories.ProposalRepository,
	logger *zap.SugaredLogger,
) commands.Command {
	return &pendingProposalsCommand{
		userRepository:     userRepository,
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *pendingProposalsCommand) CanHandle(command string) bool {
	return command == pendingProposalsCommandName
}

func (c *pendingProposalsCommand) Handle(command, arguments string, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable {
	proposals, err := c.proposalRepository.GetManyByStatus(models.ProposalStatusCreated)
	if err != nil {
		c.logger.Errorw("failed to get proposals", "error", err)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}

	user.TelegramState.LastCommand = ""

	_, err = c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
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

		var message tgbotapi.MessageConfig

		switch user.Role {
		case models.UserRoleMember:
			if proposal.NominatorID != user.ID {
				messageText += fmt.Sprintf(
					"\n%s",
					"Если ты тоже считаешь, что этот человек должен скорее стать частью сообщества, нажми на кнопку *Оставить комментарий* и сформулируй — почему ты так считаешь? Дополнительные мнения помогают в процессе отбора.",
				)
			}

			message = tgbotapi.NewMessage(chatID, messageText)
			message.ParseMode = tgbotapi.ModeMarkdown
			message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Оставить комментарий", fmt.Sprintf("add_comment:%d", proposal.ID)),
				),
			)
		case models.UserRoleSeeder:
			pollChatID := strings.TrimPrefix(strconv.Itoa(proposal.Poll.ChatID), "-100")

			message = tgbotapi.NewMessage(chatID, messageText)
			message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("Проголосовать", fmt.Sprintf("https://t.me/c/%s/%d", pollChatID, proposal.Poll.PollMessageID)),
					tgbotapi.NewInlineKeyboardButtonURL("Обсудить", fmt.Sprintf("https://t.me/c/%s/%d", pollChatID, proposal.Poll.DiscussionMessageID)),
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
