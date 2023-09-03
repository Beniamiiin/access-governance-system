package repositories

import (
	"access_governance_system/internal/db/models"

	"github.com/go-pg/pg/v10"
)

type userRepository struct {
	repository
}

type UserRepository interface {
	Create(request *models.User) (*models.User, error)
	Update(request *models.User) (*models.User, error)
	GetOneByTelegramID(telegramID int64) (*models.User, error)
	GetOneByTelegramNickname(telegramNickname string) (*models.User, error)
	GetMany() ([]*models.User, error)
}

func NewUserRepository(db *pg.DB) UserRepository {
	return &userRepository{
		repository: repository{
			db: db,
		},
	}
}

func (r *userRepository) Create(request *models.User) (*models.User, error) {
	_, err := r.db.Model(request).Insert()
	if err != nil {
		return nil, err
	}

	user := &models.User{}

	err = r.db.Model(user).
		Relation("Proposals").
		Where("id = ?", request.ID).
		Select()

	return user, err
}

func (r *userRepository) Update(request *models.User) (*models.User, error) {
	_, err := r.db.Model(request).WherePK().Update()
	if err != nil {
		return nil, err
	}

	user := &models.User{}

	err = r.db.Model(user).
		Relation("Proposals").
		Where("id = ?", request.ID).
		Select()

	return user, err
}

func (r *userRepository) GetOneByTelegramID(telegramID int64) (*models.User, error) {
	user := &models.User{}

	err := r.db.Model(user).
		Relation("Proposals").
		Where("telegram_id = ?", telegramID).
		Select()

	return user, err
}

func (r *userRepository) GetOneByTelegramNickname(telegramNickname string) (*models.User, error) {
	user := &models.User{}

	err := r.db.Model(user).
		Relation("Proposals").
		Where("telegram_nickname = ?", telegramNickname).
		Select()

	return user, err
}

func (r *userRepository) GetMany() ([]*models.User, error) {
	users := make([]*models.User, 0)

	err := r.db.Model(&users).
		Relation("Proposals").
		Select()

	return users, err
}
