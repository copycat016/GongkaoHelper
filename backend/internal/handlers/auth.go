package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type AuthHandler struct {
	service *services.AuthService
}

type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var payload loginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40010, "invalid login payload")
		return
	}

	result, err := h.service.Login(payload.Username, payload.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, 40101, "invalid username or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50010, "login failed")
		return
	}
	response.Success(c, result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	value, exists := c.Get("auth_user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40102, "unauthorized")
		return
	}

	user, ok := value.(models.User)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40102, "unauthorized")
		return
	}

	response.Success(c, h.service.UserInfo(user))
}
