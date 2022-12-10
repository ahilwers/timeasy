package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"gorm.io/gorm"
)

type gormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(database *gorm.DB) repository.UserRepository {
	return &gormUserRepository{
		db: database,
	}
}

func (repo *gormUserRepository) AddUser(user *model.User) (*model.User, error) {
	if err := repo.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
