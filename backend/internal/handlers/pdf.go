package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type PDFHandler struct{}

func NewPDFHandler() *PDFHandler {
	return &PDFHandler{}
}

func (h *PDFHandler) ParseTest(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40081, "pdf file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		response.Error(c, http.StatusBadRequest, 40082, "file must be a pdf")
		return
	}

	if err := os.MkdirAll(filepath.Join(os.TempDir(), "gkweb-pdf-test"), 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, 50081, "prepare temp directory failed")
		return
	}
	tempPath := filepath.Join(os.TempDir(), "gkweb-pdf-test", fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	defer os.Remove(tempPath)

	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		response.Error(c, http.StatusInternalServerError, 50082, "save temp pdf failed")
		return
	}

	pages, quality, err := services.ExtractPDFTextPagesForTest(tempPath)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40083, err.Error())
		return
	}

	totalChars := 0
	for _, page := range pages {
		totalChars += len([]rune(page.Text))
	}

	response.Success(c, gin.H{
		"file_name":   file.Filename,
		"page_count":  len(pages),
		"total_chars": totalChars,
		"quality":     quality,
		"pages":       pages,
	})
}
