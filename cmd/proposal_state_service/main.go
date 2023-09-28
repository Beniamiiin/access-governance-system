package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/di"
	"access_governance_system/internal/services"
	"fmt"
	"math"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func main() {
	config, err := configs.LoadProposalStateServiceConfig()
	logger := di.NewLogger(config.Logger.AppName, config.App.Environment, config.Logger.URL)

	if err != nil {
		logger.Fatalw("failed to load config", "error", err)
	}
	logger.Info("config loaded")

	logger.Info("starting db")
	database, err := db.StartDB(config.DB, logger)
	if err != nil {
		logger.Fatalw("failed to start db", "error", err)
	}
	logger.Info("db started")

	logger.Info("initializing repositories and services")
	userRepository := repositories.NewUserRepository(database)
	proposalRepository := repositories.NewProposalRepository(database)
	voteService := services.NewVoteService(config.VoteAPI.URL)

	logger.Info("getting seeders")
	seeders, err := userRepository.GetManyByRole(models.UserRoleSeeder)
	if err != nil {
		logger.Fatalw("failed to get seeders", "error", err)
	}

	logger.Info("getting proposals")
	proposals, err := proposalRepository.GetManyByStatus(models.ProposalStatusCreated)
	if err != nil {
		logger.Fatalw("failed to get proposals", "error", err)
	}

	proposalsNeedToBeUpdated := getProposalsNeedToBeUpdated(seeders, proposals, userRepository, voteService, config, logger)

	if len(proposalsNeedToBeUpdated) == 0 {
		logger.Info("no proposals to update")
	} else {
		updatedProposals := updateProposals(proposalsNeedToBeUpdated, proposalRepository, logger)

		for _, proposal := range updatedProposals {
			sendNotifications(proposal, userRepository, config, logger)
		}

		logger.Info("proposals updated")
	}
}

func getProposalsNeedToBeUpdated(
	seeders []*models.User,
	proposals []*models.Proposal,
	userRepository repositories.UserRepository,
	voteService services.VoteService,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) []*models.Proposal {
	minRequiredSeedersCount := int(math.Round(float64(len(seeders)) * config.Quorum))

	var proposalsNeedToBeUpdated []*models.Proposal

	for _, proposal := range proposals {
		if !compareDatesWithoutTime(proposal.FinishedAt, time.Now()) {
			continue
		}

		votes, err := voteService.GetVotes(proposal.Poll.ID)
		if err != nil {
			logger.Errorw("failed to get votes", "error", err)
			continue
		}

		yesVotes, noVotes := calculateYesAndNoVotes(votes, config.YesVotesToOvercomeNo)
		votedUsers := getVotedUsers(votes, userRepository, logger)
		votedSeeders := getVotedSeeders(votedUsers)

		updateProposalStatus(proposal, votes, votedSeeders, minRequiredSeedersCount, yesVotes, noVotes, config)

		proposalsNeedToBeUpdated = append(proposalsNeedToBeUpdated, proposal)
	}

	return proposalsNeedToBeUpdated
}

func calculateYesAndNoVotes(votes []services.Vote, yesVotesToOvercomeNo float64) (yesVotes int, noVotes int) {
	for _, vote := range votes {
		if vote.Option == "yes" {
			yesVotes++
		} else if vote.Option == "no" {
			noVotes++
		}
	}

	minRequiredYesVotes := int(math.Round(float64(len(votes)) * yesVotesToOvercomeNo))

	if yesVotes >= minRequiredYesVotes && noVotes > 0 {
		noVotes--
	}

	return yesVotes, noVotes
}

func getVotedUsers(votes []services.Vote, userRepository repositories.UserRepository, logger *zap.SugaredLogger) []*models.User {
	votedUsers := []*models.User{}

	for _, vote := range votes {
		user, err := userRepository.GetOneByTelegramNickname(vote.Username)
		if err != nil {
			logger.Errorw("failed to get user", "error", err)
			continue
		} else if user == nil {
			logger.Errorw("user not found", "username", vote.Username)
			continue
		}

		votedUsers = append(votedUsers, user)
	}

	return votedUsers
}

func getVotedSeeders(users []*models.User) []*models.User {
	votedSeeders := []*models.User{}

	for _, user := range users {
		if user.Role == models.UserRoleSeeder {
			votedSeeders = append(votedSeeders, user)
		}
	}

	return votedSeeders
}

func updateProposalStatus(
	proposal *models.Proposal,
	votes []services.Vote,
	votedSeeders []*models.User,
	minRequiredSeedersCount,
	yesVotes,
	noVotes int,
	config configs.ProposalStateServiceConfig,
) {
	proposal.Status = models.ProposalStatusApproved

	minRequiredYesVotes := int(math.Round(float64(len(votes)) * config.MinYesPercentage))

	if len(votedSeeders) < minRequiredSeedersCount {
		proposal.Status = models.ProposalStatusNoQuorum
	} else if yesVotes < minRequiredYesVotes || noVotes > 0 {
		proposal.Status = models.ProposalStatusRejected
	}
}

func updateProposals(
	proposals []*models.Proposal,
	proposalRepository repositories.ProposalRepository,
	logger *zap.SugaredLogger,
) []*models.Proposal {
	var updatedProposals []*models.Proposal

	for _, proposal := range proposals {
		_, err := proposalRepository.Update(proposal)

		if err != nil {
			logger.Errorw("failed to update proposal", "error", err)
		} else {
			updatedProposals = append(updatedProposals, proposal)
		}
	}

	return updatedProposals
}

func sendNotifications(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	switch proposal.Status {
	case models.ProposalStatusRejected:
		sendNotificationsIfProposalRejected(proposal, userRepository, config, logger)
	case models.ProposalStatusApproved:
		sendNotificationsIfProposalApproved(proposal, userRepository, config, logger)
	case models.ProposalStatusNoQuorum:
		sendNotificationsIfProposalNoQuourum(proposal, userRepository, config, logger)
	}
}

func sendNotificationsIfProposalRejected(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	nominator, err := userRepository.GetOneByID(proposal.NominatorID)
	if err != nil {
		logger.Fatalf("could not get nominator: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		logger.Fatalf("could not create bot: %v", err)
	}

	messages := []tgbotapi.MessageConfig{
		messageForProposalRejectedToNominator(proposal, nominator),
		messageForProposalRejectedToSeedersGroup(proposal),
	}

	for _, message := range messages {
		_, err = bot.Send(message)
		if err != nil {
			logger.Fatalf("could not send message: %v", err)
		}
	}
}

func messageForProposalRejectedToNominator(proposal *models.Proposal, nominator *models.User) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(nominator.TelegramID), text)
	return message
}

func messageForProposalRejectedToSeedersGroup(proposal *models.Proposal) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(proposal.Poll.ChatID), text)
	message.BaseChat.ReplyToMessageID = proposal.Poll.PollMessageID
	return message
}

func sendNotificationsIfProposalApproved(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	nominator, err := userRepository.GetOneByID(proposal.NominatorID)
	if err != nil {
		logger.Fatalf("could not get nominator: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		logger.Fatalf("could not create bot: %v", err)
	}

	inviteLinkConfig := tgbotapi.ChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: config.App.MembersChatID,
		},
	}
	inviteLink, err := bot.GetInviteLink(inviteLinkConfig)
	if err != nil {
		logger.Fatalf("could not get invite link: %v", err)
	}

	messages := []tgbotapi.MessageConfig{
		messageForProposalApprovedToNominator(proposal, nominator, inviteLink),
	}

	for _, message := range messages {
		_, err = bot.Send(message)
		if err != nil {
			logger.Fatalf("could not send message: %v", err)
		}
	}
}

func messageForProposalApprovedToNominator(proposal *models.Proposal, nominator *models.User, inviteLink string) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была принята. Отправь ссылку(%s) кандидату, чтобы он зашел в группу. ", proposal.NomineeName, proposal.NomineeTelegramNickname, inviteLink)
	message := tgbotapi.NewMessage(int64(nominator.TelegramID), text)
	message.DisableWebPagePreview = true
	return message
}

func sendNotificationsIfProposalNoQuourum(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	nominator, err := userRepository.GetOneByID(proposal.NominatorID)
	if err != nil {
		logger.Fatalf("could not get nominator: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		logger.Fatalf("could not create bot: %v", err)
	}

	messages := []tgbotapi.MessageConfig{
		messageForProposalNoQuourumToNominator(proposal, nominator),
		messageForProposalNoQuourumToSeedersGroup(proposal),
	}

	for _, message := range messages {
		_, err = bot.Send(message)
		if err != nil {
			logger.Fatalf("could not send message: %v", err)
		}
	}
}

func messageForProposalNoQuourumToNominator(proposal *models.Proposal, nominator *models.User) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена по причине отсутствия кворума.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(nominator.TelegramID), text)
	return message
}

func messageForProposalNoQuourumToSeedersGroup(proposal *models.Proposal) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена по причине отсутствия кворума.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(proposal.Poll.ChatID), text)
	message.BaseChat.ReplyToMessageID = proposal.Poll.PollMessageID
	return message
}

func compareDatesWithoutTime(date1 time.Time, date2 time.Time) bool {
	return date1.Year() == date2.Year() && date1.Month() == date2.Month() && date1.Day() == date2.Day()
}
