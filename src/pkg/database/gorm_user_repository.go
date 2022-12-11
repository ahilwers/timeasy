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

func (repo *gormUserRepository) GetAllUsers() ([]model.User, error) {
	var users []model.User
	if err := repo.db.Order("username").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
