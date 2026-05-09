package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		response.Error(c, http.StatusInternalServerError, 50000, "internal server error")
	})
}
