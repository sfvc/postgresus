package users

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	user_enums "postgresus-backend/internal/features/users/enums"
	user_models "postgresus-backend/internal/features/users/models"
)

type UserController struct {
	userService   *UserService
	signinLimiter *rate.Limiter
}

func (c *UserController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/users/signup", c.SignUp)
	router.POST("/users/signin", c.SignIn)
	router.GET("/users/is-any-user-exist", c.IsAnyUserExist)

	// Admin-only routes for user management
	router.POST("/users/admin/create-user", c.CreateUser)
	router.GET("/users/admin/list", c.ListUsers)
	router.PUT("/users/admin/:id/status", c.UpdateUserStatus)
	router.PUT("/users/admin/:id/password", c.ChangeUserPassword)
}

// SignUp
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "User signup data"
// @Success 200
// @Failure 400
// @Router /users/signup [post]
func (c *UserController) SignUp(ctx *gin.Context) {
	var request SignUpRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err := c.userService.SignUp(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// SignIn
// @Summary Authenticate a user
// @Description Authenticate a user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body SignInRequest true "User signin data"
// @Success 200 {object} SignInResponse
// @Failure 400
// @Failure 429 {object} map[string]string "Rate limit exceeded"
// @Router /users/signin [post]
func (c *UserController) SignIn(ctx *gin.Context) {
	// We use rate limiter to prevent brute force attacks
	if !c.signinLimiter.Allow() {
		ctx.JSON(
			http.StatusTooManyRequests,
			gin.H{"error": "Rate limit exceeded. Please try again later."},
		)
		return
	}

	var request SignInRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	response, err := c.userService.SignIn(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// IsAnyUserExist
// @Summary Check if any user exists
// @Description Check if any user exists in the system
// @Tags users
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /users/is-any-user-exist [get]
func (c *UserController) IsAnyUserExist(ctx *gin.Context) {
	isExist, err := c.userService.IsAnyUserExist()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"isExist": isExist})
}

// Helper method to get authenticated user from Authorization header
func (c *UserController) getAuthenticatedUser(ctx *gin.Context) (*user_models.User, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	// Handle both "Bearer <token>" and direct token formats
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	user, err := c.userService.GetUserFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if user is blocked
	if user.Status == user_enums.UserStatusBlocked {
		return nil, fmt.Errorf("user account is blocked")
	}

	return user, nil
}

// CreateUser
// @Summary Create a new user (Admin only)
// @Description Create a new user with specified role. Only accessible by admin users.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT token"
// @Param request body CreateUserRequest true "User creation data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/admin/create-user [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	user, err := c.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != user_enums.UserRoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only admin users can create other users"})
		return
	}

	var request CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err = c.userService.CreateUser(user, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// ListUsers
// @Summary List all users (Admin only)
// @Description Get a list of all users in the system. Only accessible by admin users.
// @Tags users
// @Produce json
// @Param Authorization header string true "JWT token"
// @Success 200 {array} UserResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/admin/list [get]
func (c *UserController) ListUsers(ctx *gin.Context) {
	user, err := c.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != user_enums.UserRoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only admin users can list all users"})
		return
	}

	users, err := c.userService.GetAllUsers(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// UpdateUserStatus
// @Summary Update user status (Admin only)
// @Description Update the status of a user (ACTIVE/BLOCKED). Only accessible by admin users.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT token"
// @Param id path string true "User ID"
// @Param request body UpdateUserStatusRequest true "Status update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/admin/{id}/status [put]
func (c *UserController) UpdateUserStatus(ctx *gin.Context) {
	user, err := c.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != user_enums.UserRoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only admin users can update user status"})
		return
	}

	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request UpdateUserStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err = c.userService.UpdateUserStatus(user, userID, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

// ChangeUserPassword
// @Summary Change user password (Admin only)
// @Description Change the password of a user. Only accessible by admin users.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT token"
// @Param id path string true "User ID"
// @Param request body ChangeUserPasswordRequest true "Password change data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/admin/{id}/password [put]
func (c *UserController) ChangeUserPassword(ctx *gin.Context) {
	user, err := c.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != user_enums.UserRoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only admin users can change other users' passwords"})
		return
	}

	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request ChangeUserPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err = c.userService.ChangeUserPassword(user, userID, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
