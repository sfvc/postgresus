package user_repositories

import (
	user_enums "postgresus-backend/internal/features/users/enums"
	user_models "postgresus-backend/internal/features/users/models"
	"postgresus-backend/internal/storage"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct{}

func (r *UserRepository) IsAnyUserExist() (bool, error) {
	var user user_models.User

	if err := storage.GetDb().First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *UserRepository) CreateUser(user *user_models.User) error {
	return storage.GetDb().Create(user).Error
}

func (r *UserRepository) GetUserByEmail(email string) (*user_models.User, error) {
	var user user_models.User
	if err := storage.GetDb().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(userID string) (*user_models.User, error) {
	var user user_models.User

	if err := storage.GetDb().Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetFirstUser() (*user_models.User, error) {
	var user user_models.User

	if err := storage.GetDb().First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateUserPassword(userID uuid.UUID, hashedPassword string) error {
	return storage.GetDb().Model(&user_models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"hashed_password":        hashedPassword,
			"password_creation_time": time.Now().UTC(),
		}).Error
}

func (r *UserRepository) GetAllUsers() ([]*user_models.User, error) {
	var users []*user_models.User
	if err := storage.GetDb().Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) UpdateUserStatus(userID uuid.UUID, status user_enums.UserStatus) error {
	return storage.GetDb().Model(&user_models.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}
