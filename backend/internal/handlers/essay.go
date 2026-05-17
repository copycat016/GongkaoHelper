package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type EssayHandler struct {
	service *services.EssayService
}

type essayModelPayload struct {
	ModelID uint `json:"model_id"`
}

type essayParsePayload struct {
	RawText         string `json:"raw_text"`
	BoundaryModelID uint   `json:"boundary_model_id"`
}

type essayReviewPayload struct {
	ModelID    uint   `json:"model_id"`
	UserAnswer string `json:"user_answer"`
}

func NewEssayHandler(service *services.EssayService) *EssayHandler {
	return &EssayHandler{service: service}
}

func (h *EssayHandler) ListDocuments(c *gin.Context) {
	documents, err := h.service.ListDocuments(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50071, "list essay documents failed")
		return
	}
	response.Success(c, documents)
}

func (h *EssayHandler) DeleteDocument(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}
	if err := h.service.DeleteDocument(userIDFromRequest(c), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40076, "document not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50076, "delete essay document failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *EssayHandler) CreateDocument(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	rawText := c.PostForm("raw_text")
	documentRole := strings.TrimSpace(c.PostForm("document_role"))
	sourceGroup := strings.TrimSpace(c.PostForm("source_group"))
	boundaryModelID := uintFromString(c.PostForm("boundary_model_id"))

	var originalName string
	var relativePath string

	file, err := c.FormFile("file")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".pdf" {
			response.Error(c, http.StatusBadRequest, 40071, "essay document must be a pdf")
			return
		}
		if err := os.MkdirAll(filepath.Join("uploads", "essay"), 0755); err != nil {
			response.Error(c, http.StatusInternalServerError, 50072, "prepare essay upload directory failed")
			return
		}
		storedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		relativePath = filepath.Join("uploads", "essay", storedName)
		if err := c.SaveUploadedFile(file, relativePath); err != nil {
			response.Error(c, http.StatusInternalServerError, 50073, "save essay pdf failed")
			return
		}
		originalName = file.Filename
		if title == "" {
			title = strings.TrimSuffix(file.Filename, ext)
		}
	}

	if title == "" {
		title = "申论 PDF 结构化任务"
	}

	document := models.EssayDocument{
		BaseModel:    models.BaseModel{UserID: userIDFromRequest(c)},
		Title:        title,
		DocumentRole: documentRole,
		SourceGroup:  sourceGroup,
		OriginalName: originalName,
		FilePath:     relativePath,
		Status:       "uploaded",
	}
	if err := h.service.CreateDocument(&document); err != nil {
		response.Error(c, http.StatusInternalServerError, 50074, "create essay document failed")
		return
	}

	if strings.TrimSpace(rawText) != "" || strings.TrimSpace(relativePath) != "" {
		parsedDocument, sections, err := h.service.ParseDocumentWithBoundaryModel(userIDFromRequest(c), document.ID, rawText, boundaryModelID)
		if err != nil {
			h.service.MarkDocumentFailed(userIDFromRequest(c), document.ID, err.Error())
			response.Error(c, http.StatusBadRequest, 40075, err.Error())
			return
		}
		chunks, _ := h.service.ListChunks(userIDFromRequest(c), document.ID)
		questions, _ := h.service.ListQuestions(userIDFromRequest(c), document.ID)
		response.Success(c, gin.H{"document": parsedDocument, "sections": sections, "chunks": chunks, "questions": questions})
		return
	}

	response.Success(c, document)
}

func (h *EssayHandler) ParseDocument(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}

	var payload essayParsePayload
	_ = c.ShouldBindJSON(&payload)

	document, sections, err := h.service.ParseDocumentWithBoundaryModel(userIDFromRequest(c), id, payload.RawText, payload.BoundaryModelID)
	if err != nil {
		h.service.MarkDocumentFailed(userIDFromRequest(c), id, err.Error())
		response.Error(c, http.StatusBadRequest, 40075, err.Error())
		return
	}
	chunks, _ := h.service.ListChunks(userIDFromRequest(c), id)
	questions, _ := h.service.ListQuestions(userIDFromRequest(c), id)
	response.Success(c, gin.H{"document": document, "sections": sections, "chunks": chunks, "questions": questions})
}

func (h *EssayHandler) DebugBoundary(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}

	var payload essayParsePayload
	_ = c.ShouldBindJSON(&payload)

	result, err := h.service.DebugBoundarySplit(userIDFromRequest(c), id, payload.RawText, payload.BoundaryModelID)
	if err != nil {
		writeServiceError(c, err, "debug essay boundary failed")
		return
	}
	response.Success(c, result)
}

func (h *EssayHandler) ListSections(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}
	sections, err := h.service.ListSections(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "list essay sections failed")
		return
	}
	response.Success(c, sections)
}

func (h *EssayHandler) ListChunks(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}
	chunks, err := h.service.ListChunks(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "list essay chunks failed")
		return
	}
	response.Success(c, chunks)
}

func (h *EssayHandler) ClassifyChunks(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}

	var payload essayModelPayload
	_ = c.ShouldBindJSON(&payload)

	chunks, err := h.service.ClassifyChunks(userIDFromRequest(c), id, payload.ModelID)
	if err != nil {
		writeServiceError(c, err, "classify essay chunks failed")
		return
	}
	response.Success(c, chunks)
}

func (h *EssayHandler) AssembleQuestions(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}
	questions, err := h.service.AssembleQuestions(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "assemble essay questions failed")
		return
	}
	response.Success(c, questions)
}

func (h *EssayHandler) ListQuestions(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, "invalid document id")
		return
	}
	questions, err := h.service.ListQuestions(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "list essay questions failed")
		return
	}
	response.Success(c, questions)
}

func (h *EssayHandler) ReviewAnswer(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40073, "invalid question id")
		return
	}

	var payload essayReviewPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40074, "invalid review payload")
		return
	}

	result, err := h.service.ReviewAnswer(userIDFromRequest(c), id, payload.ModelID, payload.UserAnswer)
	if err != nil {
		writeServiceError(c, err, "review essay answer failed")
		return
	}
	response.Success(c, result)
}
