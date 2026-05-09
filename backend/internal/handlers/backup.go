package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type BackupHandler struct {
	service *services.BackupService
}

func NewBackupHandler(service *services.BackupService) *BackupHandler {
	return &BackupHandler{service: service}
}

func (h *BackupHandler) Export(c *gin.Context) {
	includeSecrets := c.Query("include_secrets") == "true"
	backup, err := h.service.Export(includeSecrets)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50071, "export backup failed")
		return
	}

	fileName := fmt.Sprintf("gkweb-backup-%s.json", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", `attachment; filename="`+fileName+`"`)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, backup)
}
