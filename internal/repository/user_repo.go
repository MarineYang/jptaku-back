package repository

import (
	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Settings").Preload("Onboarding").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) CreateSettings(settings *model.UserSettings) error {
	return r.db.Create(settings).Error
}

func (r *UserRepository) UpdateSettings(settings *model.UserSettings) error {
	return r.db.Save(settings).Error
}

func (r *UserRepository) GetSettings(userID uint) (*model.UserSettings, error) {
	var settings model.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *UserRepository) CreateOnboarding(onboarding *model.UserOnboarding) error {
	return r.db.Create(onboarding).Error
}

func (r *UserRepository) UpdateOnboarding(onboarding *model.UserOnboarding) error {
	return r.db.Save(onboarding).Error
}

func (r *UserRepository) GetOnboarding(userID uint) (*model.UserOnboarding, error) {
	var onboarding model.UserOnboarding
	err := r.db.Where("user_id = ?", userID).First(&onboarding).Error
	if err != nil {
		return nil, err
	}
	return &onboarding, nil
}
