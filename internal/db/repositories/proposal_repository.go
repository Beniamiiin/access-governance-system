package repositories

import (
	"access_governance_system/internal/db/models"
	"errors"

	"github.com/go-pg/pg/v10"
)

type proposalRepository struct {
	repository
}

type ProposalRepository interface {
	Create(request *models.Proposal) (*models.Proposal, error)
	Update(request *models.Proposal) (*models.Proposal, error)
	Delete(request *models.Proposal) error
	GetOneByID(id int64) (*models.Proposal, error)
	GetManyByNomineeNickname(nomineeNickName string) ([]*models.Proposal, error)
	GetApprovedByNomineeNickname(nomineeNickName string) (*models.Proposal, error)
	GetManyByStatus(status ...models.ProposalStatus) ([]*models.Proposal, error)
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
		Where("id = ?", request.ID).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return proposal, err
}

func (r *proposalRepository) Update(request *models.Proposal) (*models.Proposal, error) {
	_, err := r.db.Model(request).WherePK().Update()
	if err != nil {
		return nil, err
	}

	proposal := &models.Proposal{}

	err = r.db.Model(proposal).
		Where("id = ?", request.ID).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return proposal, err
}

func (r *proposalRepository) Delete(request *models.Proposal) error {
	_, err := r.db.Model(request).WherePK().Delete()
	return err
}

func (r *proposalRepository) GetOneByID(id int64) (*models.Proposal, error) {
	proposal := &models.Proposal{}

	err := r.db.Model(proposal).
		Where("id = ?", id).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return proposal, err
}

func (r *proposalRepository) GetManyByNomineeNickname(nomineeNickName string) ([]*models.Proposal, error) {
	proposals := make([]*models.Proposal, 0)

	err := r.db.Model(&proposals).
		Where("nominee_telegram_nickname = ?", nomineeNickName).
		OrderExpr("created_at ASC").
		Select()

	return proposals, err
}

func (r *proposalRepository) GetApprovedByNomineeNickname(nomineeNickName string) (*models.Proposal, error) {
	proposals := make([]*models.Proposal, 0)

	err := r.db.Model(&proposals).
		Where("nominee_telegram_nickname = ? AND status = ?", nomineeNickName, models.ProposalStatusApproved).
		OrderExpr("created_at ASC").
		Select()

	if len(proposals) == 0 {
		return nil, errors.New("proposal not found")
	}

	return proposals[len(proposals)-1], err
}

func (r *proposalRepository) GetManyByStatus(status ...models.ProposalStatus) ([]*models.Proposal, error) {
	proposals := make([]*models.Proposal, 0)

	err := r.db.Model(&proposals).
		WhereGroup(func(q *pg.Query) (*pg.Query, error) {
			for _, s := range status {
				q = q.WhereOr("status = ?", s)
			}
			return q, nil
		}).
		Select()

	return proposals, err
}
