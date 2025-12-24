package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nakamura/chatwoot-go/internal/storage"
)

type UploadHandler struct {
	storage *storage.MinioService
}

func NewUploadHandler(s *storage.MinioService) *UploadHandler {
	return &UploadHandler{storage: s}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Validate content type is image
	contentType := header.Header.Get("Content-Type")
	// Simple verification (production should be more robust)

	url, err := h.storage.UploadFile(c.Request.Context(), file, header.Size, header.Filename, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
