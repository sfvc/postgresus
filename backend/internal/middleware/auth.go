package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"postgresus-backend/internal/features/users"
	user_enums "postgresus-backend/internal/features/users/enums"
	users_models "postgresus-backend/internal/features/users/models"
)

type AuthMiddleware struct {
	userService *users.UserService
}

func NewAuthMiddleware(userService *users.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// RequireAuth validates JWT token and sets user in context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := m.extractUserFromRequest(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			c.Abort()
			return
		}

		// Check if user is blocked
		if user.Status == user_enums.UserStatusBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "User account is blocked"})
			c.Abort()
			return
		}

		// Set user in context for downstream handlers
		c.Set("user", user)
		c.Next()
	}
}

// RequireRole validates that the authenticated user has the specified role
func (m *AuthMiddleware) RequireRole(requiredRole user_enums.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error: user not found in context"})
			c.Abort()
			return
		}

		userModel, ok := user.(*users_models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error: invalid user type"})
			c.Abort()
			return
		}

		if userModel.Role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole validates that the authenticated user has any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(roles ...user_enums.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error: user not found in context"})
			c.Abort()
			return
		}

		userModel, ok := user.(*users_models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error: invalid user type"})
			c.Abort()
			return
		}

		for _, role := range roles {
			if userModel.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// AdminOnly is a convenience method for requiring admin role
func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return m.RequireRole(user_enums.UserRoleAdmin)
}

// extractUserFromRequest extracts and validates the user from the Authorization header
func (m *AuthMiddleware) extractUserFromRequest(c *gin.Context) (*users_models.User, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, &AuthError{Message: "Authorization header is required"}
	}

	// Handle both "Bearer <token>" and direct token formats
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if token == "" {
		return nil, &AuthError{Message: "Token is required"}
	}

	user, err := m.userService.GetUserFromToken(token)
	if err != nil {
		return nil, &AuthError{Message: "Invalid token"}
	}

	return user, nil
}

// GetUserFromContext is a helper function to extract user from gin context
func GetUserFromContext(c *gin.Context) (*users_models.User, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, &AuthError{Message: "User not found in context"}
	}

	userModel, ok := user.(*users_models.User)
	if !ok {
		return nil, &AuthError{Message: "Invalid user type in context"}
	}

	return userModel, nil
}

// AuthError represents authentication/authorization errors
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
