package agbcommands

import (
	"fmt"
	"strings"
	"time"

	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/services"
	"access_governance_system/internal/tg_bot/commands"
	tgbot "access_governance_system/internal/tg_bot/extension"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	createProposalCommandName = "create_proposal"

	waitingForTypeState     = "waiting_for_type"
	waitingForNicknameState = "waiting_for_nickname"
	waitingForNameState     = "waiting_for_name"
	waitingForReasonState   = "waiting_for_reason"
	waitingForConfirmState  = "waiting_for_confirm"
)

var (
	proposalTypeMember = models.UserRoleMember.String()
	proposalTypeSeeder = models.UserRoleSeeder.String()

	confirmYes = "Да"
	confirmNo  = "Нет, начать заново"
)

type createProposalCommand struct {
	config             configs.AccessGovernanceBotConfig
	userRepository     repositories.UserRepository
	proposalRepository repositories.ProposalRepository
	voteService        services.VoteService

	logger *zap.SugaredLogger
}

func NewCreateProposalCommand(
	config configs.AccessGovernanceBotConfig,
	userRepository repositories.UserRepository,
	proposalRepository repositories.ProposalRepository,
	voteService services.VoteService,

	logger *zap.SugaredLogger,
) commands.Command {
	return &createProposalCommand{
		config:             config,
		userRepository:     userRepository,
		proposalRepository: proposalRepository,
		voteService:        voteService,

		logger: logger,
	}
}

func (c *createProposalCommand) CanHandle(command string) bool {
	return command == createProposalCommandName
}

func (c *createProposalCommand) Handle(
	command, arguments string,
	user *models.User,
	bot *tgbotapi.BotAPI,
	chatID int64,
) []tgbotapi.Chattable {
	var message tgbotapi.Chattable

	if command == createProposalCommandName {
		message = c.handleCreateProposalCommand(user, chatID)
	} else {
		switch user.TelegramState.LastCommandState {
		case waitingForTypeState:
			message = c.handleWaitingForTypeState(command, user, chatID)
		case waitingForNicknameState:
			message = c.handleWaitingForNicknameState(command, user, chatID)
		case waitingForNameState:
			message = c.handleWaitingForNameState(command, user, chatID)
		case waitingForReasonState:
			message = c.handleWaitingForReasonState(command, user, chatID)
		case waitingForConfirmState:
			message = c.handleWaitingForConfirmState(command, user, chatID)

			if command == confirmYes {
				var text string
				switch user.TempProposal.NomineeRole {
				case models.NomineeRoleMember:
					text = fmt.Sprintf(
						"@%s предлагает добавить @%s в сообщество",
						user.TelegramNickname,
						user.TempProposal.NomineeTelegramNickname,
					)
				case models.NomineeRoleSeeder:
					text = fmt.Sprintf(
						"@%s предлагает повысить @%s до seeder",
						user.TelegramNickname,
						user.TempProposal.NomineeTelegramNickname,
					)
				}

				_, err := bot.Send(tgbotapi.NewMessage(c.config.App.MembersChatID, text))
				if err != nil {
					c.logger.Errorw("could not send message", "error", err)
				}

				user.TempProposal = models.Proposal{}
				user.TelegramState = models.TelegramState{}
				_ = c.updateUser(user)
			}
		default:
			c.logger.Errorw("user has unknown state", "error", user.TelegramState.LastCommandState)
			message = tgbot.DefaultErrorMessage(chatID)
		}
	}

	return []tgbotapi.Chattable{message}
}

func (c *createProposalCommand) handleCreateProposalCommand(user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal = models.Proposal{NominatorID: user.ID}

	switch user.Role {
	case models.UserRoleMember:
		return c.handleWaitingForTypeState(models.UserRoleMember.String(), user, chatID)
	case models.UserRoleSeeder:
		message := tgbotapi.NewMessage(chatID, "Кого ты хочешь добавить — *member* или *seeder*?")
		message.ParseMode = tgbotapi.ModeMarkdown
		message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(proposalTypeMember, proposalTypeMember),
				tgbotapi.NewInlineKeyboardButtonData(proposalTypeSeeder, proposalTypeSeeder),
			),
		)

		user.TelegramState.LastCommandState = waitingForTypeState
		_ = c.updateUser(user)

		return message
	}

	c.logger.Errorw("user has unknown role", "role", user.Role)
	return nil
}

func (c *createProposalCommand) handleWaitingForTypeState(
	proposalNomineeType string,
	user *models.User,
	chatID int64,
) tgbotapi.Chattable {
	var text string

	switch strings.ToLower(proposalNomineeType) {
	case proposalTypeMember:
		user.TempProposal.NomineeRole = models.NomineeRoleMember

		text = fmt.Sprintf(
			"Напиши никнейм пользователя *%s* в telegram в формате @nickname, которого ты хочешь добавить в сообщество. "+
				"Если у пользователя нет никнейма, то попроси его создать, так как без него мы не сможем добавить его в сообщество.",
			user.TempProposal.NomineeRole.String(),
		)
	case proposalTypeSeeder:
		user.TempProposal.NomineeRole = models.NomineeRoleSeeder

		text = fmt.Sprintf(
			"Напиши никнейм пользователя *%s* в telegram в формате @nickname, которого ты хочешь сделать сидером.",
			user.TempProposal.NomineeRole.String(),
		)
	default:
		c.logger.Warnf("user has unknown nominee type: %s", proposalNomineeType)
		return tgbotapi.NewMessage(chatID, fmt.Sprintf("Неизвестный тип участника: %s.", proposalNomineeType))
	}

	user.TelegramState.LastCommandState = waitingForNicknameState
	_ = c.updateUser(user)

	message := tgbotapi.NewMessage(chatID, text)
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func (c *createProposalCommand) handleWaitingForNicknameState(
	proposalNomineeNickname string,
	user *models.User,
	chatID int64,
) tgbotapi.Chattable {
	proposalNomineeNickname = strings.TrimPrefix(proposalNomineeNickname, "@")

	proposals, err := c.proposalRepository.GetManyByNomineeNickname(proposalNomineeNickname)
	if err != nil {
		c.logger.Errorw("failed to get proposals by nominee nickname", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	} else if len(proposals) > 0 {
		lastProposal := proposals[len(proposals)-1]

		switch lastProposal.Status {
		case models.ProposalStatusCreated:
			c.logger.Warnf(
				"user tried to create proposal for nominee with existing created proposal: %s, %d, %s",
				proposalNomineeNickname,
				lastProposal.ID,
				lastProposal.CreatedAt,
			)
			return tgbotapi.NewMessage(
				chatID,
				"Предыдущее предложение на добавление этого участника в сообщество ещё не рассмотрено.",
			)
		case models.ProposalStatusRejected:
			if !lastProposal.CreatedAt.Before(time.Now().AddDate(0, -3, 0)) {
				c.logger.Warnf(
					"user tried to create proposal for nominee with existing rejected proposal: %s, %d, %s",
					proposalNomineeNickname,
					lastProposal.ID,
					lastProposal.CreatedAt,
				)

				text := "Предыдущее предложение на добавление этого участника в сообщество было отклонено менее 3-х месяцев назад. " +
					"Участник может быть предложен к добавлению не чаще, чем один раз в три месяца."
				return tgbotapi.NewMessage(chatID, text)
			}
		}
	}

	foundUser, err := c.userRepository.GetOneByTelegramNickname(proposalNomineeNickname)
	if err != nil {
		c.logger.Errorw("failed to get user by nominee nickname", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	} else if foundUser != nil {
		if (foundUser.Role == models.UserRoleMember && user.TempProposal.NomineeRole == models.NomineeRoleMember) ||
			foundUser.Role == models.UserRoleSeeder {
			c.logger.Warnf(
				"user tried to create proposal for nominee with existing approved proposal: %s",
				proposalNomineeNickname,
			)
			return tgbotapi.NewMessage(chatID, "Этот участник уже состоит в сообществе.")
		}
	} else if foundUser == nil && user.TempProposal.NomineeRole == models.NomineeRoleSeeder {
		return tgbotapi.NewMessage(
			chatID,
			"К сожалению, я не нашел пользователя с таким никнеймом в сообществе.",
		)
	}

	var text string

	switch user.TempProposal.NomineeRole {
	case models.NomineeRoleMember:
		text = fmt.Sprintf(
			`
Проверь, что ты правильно написал никнейм пользователя: @%s, ты всегда можешь начать сначала, вызвав команду /cancel_proposal.

Если все корректно, то напиши имя и фамилию человека, которого ты хочешь добавить.
`, proposalNomineeNickname,
		)

		user.TelegramState.LastCommandState = waitingForNameState
	case models.NomineeRoleSeeder:
		text = fmt.Sprintf(
			`
Проверь, что ты правильно написал никнейм пользователя: @%s, ты всегда можешь начать сначала, вызвав команду /cancel_proposal.

Если все корректно, то напиши, почему ты считаешь, что этого человека стоит повысить до seeder? Чем подробнее описание, тем легче будет принято решение.
			`, proposalNomineeNickname,
		)

		user.TelegramState.LastCommandState = waitingForReasonState
	}

	message := tgbotapi.NewMessage(chatID, text)

	user.TempProposal.NomineeTelegramNickname = proposalNomineeNickname

	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForNameState(
	proposalNomineeName string,
	user *models.User,
	chatID int64,
) tgbotapi.Chattable {
	message := tgbotapi.NewMessage(
		chatID,
		`Теперь напиши, почему ты считаешь, что этого человека стоит добавить в сообщество? Чем подробнее описание, тем легче будет принято решение.

_В Shmit16 нет чеклиста и нет простого ответа на вопрос, кем надо быть или что надо сделать, чтобы к нам попасть. Должно сложиться так, что участники сообщества чувствуют удовольствие от общения с новым человеком и органически хотят проводить время вместе. Сообщество выросло из группы IT-предпринимателей, и за 10 лет стало шире проф ролей и приветствует любые проявления.

Важно: у нас не предусмотрен механизм исключения из сообщества, поэтому каждый, кого мы добавляем — заходит к нам в дом. 

Оформляя заявку, ты приглашаешь человека быть с тобой на фестивалях, в путешествиях, на ретритах и у тебя в гостях. Представь этого человека на наших мероприятиях и реши, будет ли классно ему с нами, и нам — с ним._`,
	)
	message.ParseMode = tgbotapi.ModeMarkdown

	user.TempProposal.NomineeName = proposalNomineeName

	user.TelegramState.LastCommandState = waitingForReasonState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForReasonState(
	proposalDescription string,
	user *models.User,
	chatID int64,
) tgbotapi.Chattable {
	user.TempProposal.Comment = proposalDescription

	text := fmt.Sprintf(
		`
Тип: *%s*
Участник: *%s (@%s)*
Комментарий: *%s*

Все правильно, отправляем предложение на голосование?

_Голосование проходит анонимно в группе из текущих активных участников (сидеры), которые являются носителями ДНК Shmit16. Решение будет принято в течение недели._
`,
		user.TempProposal.NomineeRole,
		user.TempProposal.NomineeName,
		user.TempProposal.NomineeTelegramNickname,
		user.TempProposal.Comment,
	)

	message := tgbotapi.NewMessage(chatID, text)
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(confirmYes, confirmYes),
			tgbotapi.NewInlineKeyboardButtonData(confirmNo, confirmNo),
		),
	)
	message.ParseMode = tgbotapi.ModeMarkdown

	user.TelegramState.LastCommandState = waitingForConfirmState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForConfirmState(
	confirmationState string,
	user *models.User,
	chatID int64,
) tgbotapi.Chattable {
	if confirmationState == confirmNo {
		user.TempProposal = models.Proposal{}
		user.TelegramState = models.TelegramState{LastCommand: createProposalCommandName}
		_ = c.updateUser(user)
		return c.handleCreateProposalCommand(user, chatID)
	}

	createdAt := time.Now()
	finishedAt := createdAt.AddDate(0, 0, c.config.App.VotingDurationDays)

	var description string

	switch user.TempProposal.NomineeRole {
	case models.NomineeRoleMember:
		description = fmt.Sprintf(
			"@%s предлагает добавить @%s в сообщество\n\nКомментарий: %s",
			user.TelegramNickname,
			user.TempProposal.NomineeTelegramNickname,
			user.TempProposal.Comment,
		)
	case models.NomineeRoleSeeder:
		nominee, err := c.userRepository.GetOneByTelegramNickname(user.TempProposal.NomineeTelegramNickname)
		if err != nil {
			c.logger.Errorw("failed to get nominee by telegram nickname", "error", err)
			return tgbot.DefaultErrorMessage(chatID)
		}

		user.TempProposal.NomineeName = nominee.Name

		description = fmt.Sprintf(
			"@%s предлагает повысить @%s до seeder\n\nКомментарий: %s",
			user.TelegramNickname,
			user.TempProposal.NomineeTelegramNickname,
			user.TempProposal.Comment,
		)
	}

	title := user.TempProposal.NomineeName

	dueDate := time.Date(finishedAt.Year(), finishedAt.Month(), finishedAt.Day(), 12, 0, 0, 0, finishedAt.Location())
	poll, err := c.voteService.CreatePoll(title, description, dueDate)
	if err != nil {
		c.logger.Errorw("failed to create poll", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	user.TempProposal.CreatedAt = createdAt
	user.TempProposal.FinishedAt = finishedAt

	if poll != (models.Poll{}) {
		user.TempProposal.Poll = poll
	}

	proposal, err := c.proposalRepository.Create(&user.TempProposal)
	if err != nil {
		c.logger.Errorw("failed to create proposal", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	c.logger.Info("proposal created")

	user.Proposals = append(user.Proposals, *proposal)

	if err = c.updateUser(user); err != nil {
		c.logger.Errorw("failed to update user", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}
	c.logger.Info("user updated")

	message := tgbotapi.NewMessage(chatID, "Предложение отправлено на голосование.")

	return message
}

func (c *createProposalCommand) updateUser(user *models.User) error {
	_, err := c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
	}
	return err
}
