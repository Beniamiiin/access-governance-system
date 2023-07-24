package repositories

import (
	"access_governance_system/internal/db/models"
	"github.com/go-pg/pg/v10"
)

type proposalRepository struct {
	repository
}

type ProposalRepository interface {
	Create(request *models.Proposal) (*models.Proposal, error)
	GetOne(proposalID int64) (*models.Proposal, error)
	GetMany(status models.ProposalStatus) ([]*models.Proposal, error)
}

func NewProposalRepository(db *pg.DB) ProposalRepository {
	return &proposalRepository{
		repository: repository{
			db: db,
		},
	}
}

func (r *proposalRepository) Create(request *models.Proposal) (*models.Proposal, error) {
	_, err := r.db.Model(request).Insert()
	if err != nil {
		return nil, err
	}

	proposal := &models.Proposal{}

	err = r.db.Model(proposal).
		Relation("Votes").
		Where("id = ?", request.ID).
		Select()

	return proposal, err
}

func (r *proposalRepository) GetOne(proposalID int64) (*models.Proposal, error) {
	proposal := &models.Proposal{}

	err := r.db.Model(proposal).
		Relation("Votes").
		Where("id = ?", proposalID).
		Select()

	return proposal, err
}

func (r *proposalRepository) GetMany(status models.ProposalStatus) ([]*models.Proposal, error) {
	proposals := make([]*models.Proposal, 0)

	err := r.db.Model(&proposals).
		Relation("Votes").
		Where("status = ?", status).
		Select()

	return proposals, err
}
