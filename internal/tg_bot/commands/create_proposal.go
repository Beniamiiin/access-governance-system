package commands

import (
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	createProposalCommandName = "create_proposal"

	waitingForTypeState     = "waiting_for_type"
	waitingForNicknameState = "waiting_for_nickname"
	waitingForReasonState   = "waiting_for_reason"
	waitingForConfirmState  = "waiting_for_confirm"
)

type createProposalCommand struct {
	userRepository     repositories.UserRepository
	proposalRepository repositories.ProposalRepository
	logger             *zap.SugaredLogger
}

func NewCreateProposalCommand(userRepository repositories.UserRepository, proposalRepository repositories.ProposalRepository, logger *zap.SugaredLogger) Command {
	return &createProposalCommand{
		userRepository:     userRepository,
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *createProposalCommand) CanHandle(command string) bool {
	return command == createProposalCommandName
}

func (c *createProposalCommand) Start(text string, user *models.User, chatID int64) tgbotapi.Chattable {
	if text == createProposalCommandName {
		return c.handleCreateProposalCommand(user, chatID)
	}

	switch user.TelegramState.LastCommandState {
	case waitingForTypeState:
		return c.handleWaitingForTypeState(text, user, chatID)
	case waitingForNicknameState:
		return c.handleWaitingForNicknameState(text, user, chatID)
	case waitingForReasonState:
		return c.handleWaitingForReasonState(text, user, chatID)
	case waitingForConfirmState:
		return c.handleWaitingForConfirmState(text, user, chatID)
	default:
		c.logger.Errorf("user has unknown state: %s", user.TelegramState.LastCommandState)
		return nil
	}
}

func (c *createProposalCommand) handleCreateProposalCommand(user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal = models.Proposal{NominatorID: user.ID}

	switch user.Role {
	case models.UserRoleMember:
		return c.handleWaitingForTypeState("Мембер", user, chatID)
	case models.UserRoleSeeder:
		message := tgbotapi.NewMessage(chatID, "Кого ты хочешь добавить в группу?")
		message.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Мембер"),
				tgbotapi.NewKeyboardButton("Сидер"),
			),
		)

		user.TelegramState.LastCommandState = waitingForTypeState
		_ = c.updateUser(user)

		return message
	}

	c.logger.Errorf("user has unknown role: %s", user.Role)
	return nil
}

func (c *createProposalCommand) handleWaitingForTypeState(proposalNomineeType string, user *models.User, chatID int64) tgbotapi.Chattable {
	message := tgbotapi.NewMessage(chatID, "Напиши никнейм человека, которого хочешь добавить в группу.")
	message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	switch proposalNomineeType {
	case "Мембер":
		user.TempProposal.NomineeRole = models.NomineeRoleMember
	case "Сидер":
		user.TempProposal.NomineeRole = models.NomineeRoleSeeder
	default:
		c.logger.Errorf("user has unknown nominee type: %s", proposalNomineeType)
	}

	user.TelegramState.LastCommandState = waitingForNicknameState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForNicknameState(proposalNomineeNickname string, user *models.User, chatID int64) tgbotapi.Chattable {
	message := tgbotapi.NewMessage(chatID, "Почему ты считаешь, что его стоит добавить в сообщество?")

	user.TempProposal.NomineeTelegramNickname = proposalNomineeNickname
	user.TempProposal.NomineeTelegramID = 999 // TODO: get nominee telegram id by nickname

	user.TelegramState.LastCommandState = waitingForReasonState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForReasonState(proposalDescription string, user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal.Description = proposalDescription

	message := tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"Тип: %s\nНикнейм: %s\nПричина: %s\n\nПодтвердить?",
		user.TempProposal.NomineeRole,
		user.TempProposal.NomineeTelegramNickname,
		user.TempProposal.Description,
	))
	message.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Да"),
			tgbotapi.NewKeyboardButton("Нет"),
		),
	)

	user.TelegramState.LastCommandState = waitingForConfirmState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForConfirmState(confirmationState string, user *models.User, chatID int64) tgbotapi.Chattable {
	var message tgbotapi.MessageConfig

	if confirmationState == "Нет" {
		message = tgbotapi.NewMessage(chatID, "Ну ок")
	} else {
		if _, err := c.proposalRepository.Create(&user.TempProposal); err != nil {
			c.logger.Errorw("failed to create proposal", "error", err)
			return tgbotapi.NewMessage(chatID, "Произошла ошибка, повторите попытку позже")
		}
		c.logger.Info("proposal created")

		user.Proposals = append(user.Proposals, user.TempProposal)

		if err := c.updateUser(user); err != nil {
			c.logger.Errorw("failed to update user", "error", err)
			return tgbotapi.NewMessage(chatID, "Произошла ошибка, повторите попытку позже")
		}
		c.logger.Info("user updated")

		message = tgbotapi.NewMessage(chatID, "Предложение создано")
	}

	message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	user.TempProposal = models.Proposal{}
	user.TelegramState = models.TelegramState{}
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) updateUser(user *models.User) error {
	_, err := c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
	}
	return err
}
