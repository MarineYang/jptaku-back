package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/pkg"
)

func AuthMiddleware(jwtManager *pkg.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			pkg.UnauthorizedResponse(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			pkg.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			pkg.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 유저 정보를 context에 저장
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// GetUserID는 context에서 유저 ID를 가져옵니다.
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(uint)
}
