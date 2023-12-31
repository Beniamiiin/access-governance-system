package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	mock_repositories "access_governance_system/internal/db/repositories/mocks"
	"access_governance_system/internal/services"
	mock_services "access_governance_system/internal/services/mocks"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestCalculateYesAndNoVotes_AllYesVotes(t *testing.T) {
	votes := []services.Vote{
		{Option: "yes"},
		{Option: "yes"},
	}
	yes, no := calculateYesAndNoVotes(votes, 0.5)
	assert.Equal(t, 2, yes)
	assert.Equal(t, 0, no)
}

func TestCalculateYesAndNoVotes_AllNoVotes(t *testing.T) {
	votes := []services.Vote{
		{Option: "no"},
		{Option: "no"},
	}
	yes, no := calculateYesAndNoVotes(votes, 0.5)
	assert.Equal(t, 0, yes)
	assert.Equal(t, 2, no)
}

func TestCalculateYesAndNoVotes_MixedVotes(t *testing.T) {
	votes := []services.Vote{
		{Option: "yes"},
		{Option: "no"},
		{Option: "no"},
		{Option: "no"},
	}
	yes, no := calculateYesAndNoVotes(votes, 0.5)
	assert.Equal(t, 1, yes)
	assert.Equal(t, 3, no)
}

func TestCalculateYesAndNoVotes_YesVotesToOvercomeNo(t *testing.T) {
	votes := []services.Vote{
		{Option: "yes"},
		{Option: "yes"},
		{Option: "no"},
	}
	yes, no := calculateYesAndNoVotes(votes, 0.5)
	assert.Equal(t, 2, yes)
	assert.Equal(t, 0, no)
}

func TestCalculateYesAndNoVotes_NotEnoughYesVotesToOvercomeNo(t *testing.T) {
	votes := []services.Vote{
		{Option: "yes"},
		{Option: "no"},
		{Option: "no"},
	}
	yes, no := calculateYesAndNoVotes(votes, 0.5)
	assert.Equal(t, 1, yes)
	assert.Equal(t, 2, no)
}

func TestGetVotedUsers_AllUsersFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mock_repositories.NewMockUserRepository(ctrl)
	logger := zap.NewNop().Sugar()

	votes := []services.Vote{
		{UserID: 1},
		{UserID: 2},
	}

	userRepo.EXPECT().GetOneByTelegramID(votes[0].UserID).Return(&models.User{ID: 1}, nil)
	userRepo.EXPECT().GetOneByTelegramID(votes[1].UserID).Return(&models.User{ID: 2}, nil)

	result := getVotedUsers(votes, userRepo, logger)
	assert.Equal(t, 2, len(result))
}

func TestGetVotedUsers_SomeUsersNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mock_repositories.NewMockUserRepository(ctrl)
	logger := zap.NewNop().Sugar()

	votes := []services.Vote{
		{UserID: 1},
		{UserID: 2},
	}

	userRepo.EXPECT().GetOneByTelegramID(votes[0].UserID).Return(nil, nil)
	userRepo.EXPECT().GetOneByTelegramID(votes[1].UserID).Return(&models.User{ID: 2}, nil)

	result := getVotedUsers(votes, userRepo, logger)
	assert.Equal(t, 1, len(result))
}

func TestGetVotedUsers_ErrorRetrievingSomeUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mock_repositories.NewMockUserRepository(ctrl)
	logger := zap.NewNop().Sugar()

	votes := []services.Vote{
		{UserID: 1},
		{UserID: 2},
	}

	userRepo.EXPECT().GetOneByTelegramID(votes[0].UserID).Return(nil, errors.New("database error"))
	userRepo.EXPECT().GetOneByTelegramID(votes[1].UserID).Return(&models.User{ID: 2}, nil)

	result := getVotedUsers(votes, userRepo, logger)
	assert.Equal(t, 1, len(result))
}

func TestGetVotedSeeders_AllUsersAreSeeders(t *testing.T) {
	users := []*models.User{
		{ID: 1, Role: models.UserRoleSeeder},
		{ID: 2, Role: models.UserRoleSeeder},
	}
	result := getVotedSeeders(users)
	assert.Equal(t, 2, len(result))
}

func TestGetVotedSeeders_NoUsersAreSeeders(t *testing.T) {
	users := []*models.User{
		{ID: 1, Role: models.UserRoleMember},
		{ID: 2, Role: models.UserRoleMember},
	}
	result := getVotedSeeders(users)
	assert.Equal(t, 0, len(result))
}

func TestGetVotedSeeders_MixedUserRoles(t *testing.T) {
	users := []*models.User{
		{ID: 1, Role: models.UserRoleSeeder},
		{ID: 2, Role: models.UserRoleMember},
		{ID: 3, Role: models.UserRoleSeeder},
	}
	result := getVotedSeeders(users)
	assert.Equal(t, 2, len(result))
}

func TestUpdateProposalStatus_Approved(t *testing.T) {
	config := configs.ProposalStateServiceConfig{
		MinYesPercentage: 0.5,
	}

	proposal := &models.Proposal{}
	votes := []services.Vote{{}, {}, {}}
	votedSeeders := []*models.User{{}, {}, {}}
	updateProposalStatus(proposal, votes, votedSeeders, 2, 3, 0, config)
	assert.Equal(t, models.ProposalStatusApproved, proposal.Status)
}

func TestUpdateProposalStatus_NoQuorum(t *testing.T) {
	config := configs.ProposalStateServiceConfig{
		MinYesPercentage: 0.5,
	}

	proposal := &models.Proposal{}
	votes := []services.Vote{{}, {}, {}}
	votedSeeders := []*models.User{{}}
	updateProposalStatus(proposal, votes, votedSeeders, 2, 3, 0, config)
	assert.Equal(t, models.ProposalStatusNoQuorum, proposal.Status)
}

func TestUpdateProposalStatus_RejectedLessThanMinYesVotes(t *testing.T) {
	config := configs.ProposalStateServiceConfig{
		MinYesPercentage: 0.5,
	}

	proposal := &models.Proposal{}
	votes := []services.Vote{{}, {}, {}}
	votedSeeders := []*models.User{{}, {}, {}}
	updateProposalStatus(proposal, votes, votedSeeders, 2, 1, 0, config)
	assert.Equal(t, models.ProposalStatusRejected, proposal.Status)
}

func TestUpdateProposalStatus_RejectedLessThanMinYesPercentage(t *testing.T) {
	config := configs.ProposalStateServiceConfig{
		MinYesPercentage: 0.5,
	}

	proposal := &models.Proposal{}
	votes := []services.Vote{{}, {}, {}, {}}
	votedSeeders := []*models.User{{}, {}, {}, {}}
	updateProposalStatus(proposal, votes, votedSeeders, 2, 1, 0, config)
	assert.Equal(t, models.ProposalStatusRejected, proposal.Status)
}

func TestUpdateProposalStatus_RejectedNoVotesGreaterThanZero(t *testing.T) {
	config := configs.ProposalStateServiceConfig{
		MinYesPercentage: 0.5,
	}

	proposal := &models.Proposal{}
	votes := []services.Vote{{}, {}, {}}
	votedSeeders := []*models.User{{}, {}, {}}
	updateProposalStatus(proposal, votes, votedSeeders, 2, 3, 1, config)
	assert.Equal(t, models.ProposalStatusRejected, proposal.Status)
}

func TestUpdateProposals_AllProposalsUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proposals := []*models.Proposal{
		{ID: 1, Poll: models.Poll{ID: 1}},
		{ID: 2, Poll: models.Poll{ID: 2}},
	}

	userRepo := mock_repositories.NewMockUserRepository(ctrl)
	proposalRepo := mock_repositories.NewMockProposalRepository(ctrl)
	voteService := mock_services.NewMockVoteService(ctrl)
	logger := zap.NewNop().Sugar()

	proposalRepo.EXPECT().Update(proposals[0]).Return(&models.Proposal{ID: 1}, nil)
	proposalRepo.EXPECT().Update(proposals[1]).Return(&models.Proposal{ID: 2}, nil)

	voteService.EXPECT().GetVotes(proposals[0].Poll.ID).Times(1)
	voteService.EXPECT().GetVotes(proposals[1].Poll.ID).Times(1)

	userRepo.EXPECT().Create(gomock.Any()).Times(2)

	result := updateProposals(proposals, proposalRepo, voteService, userRepo, logger)
	assert.Equal(t, 2, len(result))
}

func TestUpdateProposals_SomeProposalsNotUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proposals := []*models.Proposal{
		{ID: 1, Poll: models.Poll{ID: 1}},
		{ID: 2, Poll: models.Poll{ID: 2}},
	}

	userRepo := mock_repositories.NewMockUserRepository(ctrl)
	proposalRepo := mock_repositories.NewMockProposalRepository(ctrl)
	voteService := mock_services.NewMockVoteService(ctrl)
	logger := zap.NewNop().Sugar()

	proposalRepo.EXPECT().Update(proposals[0]).Return(nil, errors.New("database error"))
	proposalRepo.EXPECT().Update(proposals[1]).Return(&models.Proposal{ID: 2}, nil)

	voteService.EXPECT().GetVotes(proposals[0].Poll.ID).Times(0)
	voteService.EXPECT().GetVotes(proposals[1].Poll.ID).Times(1)

	userRepo.EXPECT().Create(gomock.Any()).Times(1)

	result := updateProposals(proposals, proposalRepo, voteService, userRepo, logger)
	assert.Equal(t, 1, len(result))
}

func TestUpdateProposals_NoProposalsUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proposals := []*models.Proposal{
		{ID: 1, Poll: models.Poll{ID: 1}},
		{ID: 2, Poll: models.Poll{ID: 2}},
	}

	userRepo := mock_repositories.NewMockUserRepository(ctrl)
	proposalRepo := mock_repositories.NewMockProposalRepository(ctrl)
	voteService := mock_services.NewMockVoteService(ctrl)
	logger := zap.NewNop().Sugar()

	proposalRepo.EXPECT().Update(proposals[0]).Return(nil, errors.New("database error"))
	proposalRepo.EXPECT().Update(proposals[1]).Return(nil, errors.New("database error"))

	voteService.EXPECT().GetVotes(proposals[0].Poll.ID).Times(0)
	voteService.EXPECT().GetVotes(proposals[1].Poll.ID).Times(0)

	userRepo.EXPECT().Create(gomock.Any()).Times(0)

	result := updateProposals(proposals, proposalRepo, voteService, userRepo, logger)
	assert.Equal(t, 0, len(result))
}
