package users

import (
	"time"

	"github.com/google/uuid"

	user_enums "postgresus-backend/internal/features/users/enums"
)

type SignUpRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type SignInRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignInResponse struct {
	UserID uuid.UUID `json:"userId"`
	Token  string    `json:"token"`
}

// Admin-only endpoints DTOs
type CreateUserRequest struct {
	Email    string              `json:"email"    validate:"required,email"`
	Password string              `json:"password" validate:"required,min=8"`
	Role     user_enums.UserRole `json:"role"     validate:"required"`
}

type UpdateUserStatusRequest struct {
	Status user_enums.UserStatus `json:"status" validate:"required"`
}

type ChangeUserPasswordRequest struct {
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

type ChangeMyPasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

type UserResponse struct {
	ID        uuid.UUID             `json:"id"`
	Email     string                `json:"email"`
	Role      user_enums.UserRole   `json:"role"`
	Status    user_enums.UserStatus `json:"status"`
	CreatedAt time.Time             `json:"createdAt"`
}
