package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	tokenauth "gkweb/backend/internal/auth"
	"gkweb/backend/internal/config"
	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
)

func AuthRequired(db *gorm.DB, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.Error(c, http.StatusUnauthorized, 40102, "unauthorized")
			c.Abort()
			return
		}

		claims, err := tokenauth.ParseAccessToken(token, cfg.JWTSecret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, 40103, tokenauth.TokenErrorMessage(err))
			c.Abort()
			return
		}

		var user models.User
		if err := db.Where("id = ? AND enabled = ?", claims.UserID, true).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusUnauthorized, 40102, "unauthorized")
				c.Abort()
				return
			}
			response.Error(c, http.StatusInternalServerError, 50011, "auth check failed")
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("auth_user", user)
		c.Set("username", user.Username)
		c.Next()
	}
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
