package users_models

import (
	user_enums "postgresus-backend/internal/features/users/enums"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                   uuid.UUID             `json:"id"        gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email                string                `json:"email"     gorm:"uniqueIndex;not null"`
	HashedPassword       string                `json:"-"         gorm:"not null"`
	PasswordCreationTime time.Time             `json:"-"         gorm:"not null"`
	CreatedAt            time.Time             `json:"createdAt" gorm:"not null;default:now()"`
	Role                 user_enums.UserRole   `json:"role"      gorm:"type:text;not null"`
	Status               user_enums.UserStatus `json:"status"    gorm:"type:text;not null;default:'ACTIVE'"`
}

func (User) TableName() string {
	return "users"
}
