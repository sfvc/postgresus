package users

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	user_enums "postgresus-backend/internal/features/users/enums"
	user_models "postgresus-backend/internal/features/users/models"
	user_repositories "postgresus-backend/internal/features/users/repositories"
)

type UserService struct {
	userRepository      *user_repositories.UserRepository
	secretKeyRepository *user_repositories.SecretKeyRepository
}

func (s *UserService) IsAnyUserExist() (bool, error) {
	return s.userRepository.IsAnyUserExist()
}

func (s *UserService) SignUp(request *SignUpRequest) error {
	isAnyUserExists, err := s.userRepository.IsAnyUserExist()
	if err != nil {
		return fmt.Errorf("failed to check if any user exists: %w", err)
	}

	if isAnyUserExists {
		return errors.New("admin user already registered")
	}

	existingUser, err := s.userRepository.GetUserByEmail(request.Email)
	if err == nil && existingUser != nil {
		return errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &user_models.User{
		ID:                   uuid.New(),
		Email:                request.Email,
		HashedPassword:       string(hashedPassword),
		PasswordCreationTime: time.Now().UTC(),
		CreatedAt:            time.Now().UTC(),
		Role:                 user_enums.UserRoleAdmin,
		Status:               user_enums.UserStatusActive,
	}

	if err := s.userRepository.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *UserService) SignIn(request *SignInRequest) (*SignInResponse, error) {
	user, err := s.userRepository.GetUserByEmail(request.Email)
	if err != nil {
		return nil, errors.New("user with this email does not exist")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(request.Password))
	if err != nil {
		return nil, errors.New("password is incorrect")
	}

	return s.GenerateAccessToken(user)
}

func (s *UserService) GetUserFromToken(token string) (*user_models.User, error) {
	secretKey, err := s.secretKeyRepository.GetSecretKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret key: %w", err)
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		user, err := s.userRepository.GetUserByID(userID)
		if err != nil {
			return nil, err
		}

		if passwordCreationTimeUnix, ok := claims["passwordCreationTime"].(float64); ok {
			tokenPasswordTime := time.Unix(int64(passwordCreationTimeUnix), 0)

			tokenTimeSeconds := tokenPasswordTime.Truncate(time.Second)
			userTimeSeconds := user.PasswordCreationTime.Truncate(time.Second)

			if !tokenTimeSeconds.Equal(userTimeSeconds) {
				return nil, errors.New("password has been changed, please sign in again")
			}
		} else {
			return nil, errors.New("invalid token claims: missing password creation time")
		}

		return user, nil
	}

	return nil, errors.New("invalid token")
}

func (s *UserService) ChangePassword(newPassword string) error {
	exists, err := s.userRepository.IsAnyUserExist()
	if err != nil || !exists {
		return errors.New("no users exist to change password")
	}

	user, err := s.userRepository.GetFirstUser()
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepository.UpdateUserPassword(user.ID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *UserService) GetFirstUser() (*user_models.User, error) {
	return s.userRepository.GetFirstUser()
}

func (s *UserService) GenerateAccessToken(user *user_models.User) (*SignInResponse, error) {
	secretKey, err := s.secretKeyRepository.GetSecretKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret key: %w", err)
	}

	tenYearsExpiration := time.Now().UTC().Add(time.Hour * 24 * 365 * 10)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":                  user.ID,
		"exp":                  tenYearsExpiration.Unix(),
		"iat":                  time.Now().UTC().Unix(),
		"role":                 string(user.Role),
		"passwordCreationTime": user.PasswordCreationTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &SignInResponse{
		UserID: user.ID,
		Token:  tokenString,
	}, nil
}

// Admin-only methods for user management

func (s *UserService) CreateUser(adminUser *user_models.User, request *CreateUserRequest) error {
	// Verify admin permissions
	if adminUser.Role != user_enums.UserRoleAdmin {
		return errors.New("only admin users can create other users")
	}

	// Check if user with this email already exists
	existingUser, err := s.userRepository.GetUserByEmail(request.Email)
	if err == nil && existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// Validate role
	if request.Role != user_enums.UserRoleManager && request.Role != user_enums.UserRoleAdmin {
		return errors.New("invalid role specified")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &user_models.User{
		ID:                   uuid.New(),
		Email:                request.Email,
		HashedPassword:       string(hashedPassword),
		PasswordCreationTime: time.Now().UTC(),
		CreatedAt:            time.Now().UTC(),
		Role:                 request.Role,
		Status:               user_enums.UserStatusActive,
	}

	if err := s.userRepository.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *UserService) GetAllUsers(adminUser *user_models.User) ([]*UserResponse, error) {
	// Verify admin permissions
	if adminUser.Role != user_enums.UserRoleAdmin {
		return nil, errors.New("only admin users can list all users")
	}

	users, err := s.userRepository.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	var response []*UserResponse
	for _, user := range users {
		response = append(response, &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		})
	}

	return response, nil
}

func (s *UserService) UpdateUserStatus(adminUser *user_models.User, userID uuid.UUID, request *UpdateUserStatusRequest) error {
	// Verify admin permissions
	if adminUser.Role != user_enums.UserRoleAdmin {
		return errors.New("only admin users can update user status")
	}

	// Don't allow admin to block themselves
	if adminUser.ID == userID && request.Status == user_enums.UserStatusBlocked {
		return errors.New("admin cannot block themselves")
	}

	_, err := s.userRepository.GetUserByID(userID.String())
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.userRepository.UpdateUserStatus(userID, request.Status); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (s *UserService) ChangeUserPassword(adminUser *user_models.User, userID uuid.UUID, request *ChangeUserPasswordRequest) error {
	// Verify admin permissions
	if adminUser.Role != user_enums.UserRoleAdmin {
		return errors.New("only admin users can change other users' passwords")
	}

	_, err := s.userRepository.GetUserByID(userID.String())
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepository.UpdateUserPassword(userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
