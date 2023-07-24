package repositories

import (
	"access_governance_system/internal/db/models"
	"github.com/go-pg/pg/v10"
)

type voteRepository struct {
	repository
}

type VoteRepository interface {
	Create(request *models.Vote) (*models.Vote, error)
	GetOne(voteID int64) (*models.Vote, error)
	GetMany() ([]*models.Vote, error)
}

func NewVoteRepository(db *pg.DB) VoteRepository {
	return &voteRepository{
		repository: repository{
			db: db,
		},
	}
}

func (r *voteRepository) Create(request *models.Vote) (*models.Vote, error) {
	_, err := r.db.Model(request).Insert()
	if err != nil {
		return nil, err
	}

	vote := &models.Vote{}

	err = r.db.Model(vote).
		Where("id = ?", request.ID).
		Select()

	return vote, err
}

func (r *voteRepository) GetOne(voteID int64) (*models.Vote, error) {
	vote := &models.Vote{}

	err := r.db.Model(vote).
		Where("id = ?", voteID).
		Select()

	return vote, err
}

func (r *voteRepository) GetMany() ([]*models.Vote, error) {
	votes := make([]*models.Vote, 0)

	err := r.db.Model(&votes).
		Select()

	return votes, err
}
