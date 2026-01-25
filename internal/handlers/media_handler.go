package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/alfafaa/alfafaa-blog/internal/config"
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	mediaService services.MediaService
	uploadConfig config.UploadConfig
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(mediaService services.MediaService, uploadConfig config.UploadConfig) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
		uploadConfig: uploadConfig,
	}
}

// UploadMedia handles file upload
// @Summary Upload media
// @Description Upload a file (image only, requires authentication)
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Param alt_text formData string false "Alt text for the image"
// @Success 201 {object} utils.Response{data=dto.MediaResponse} "File uploaded successfully"
// @Failure 400 {object} utils.Response "No file provided or invalid file type"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 413 {object} utils.Response "File too large"
// @Router /media/upload [post]
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "NO_FILE", "No file provided", nil)
		return
	}

	// Validate file
	if validationErr := utils.ValidateImageFile(file, h.uploadConfig.MaxSize); validationErr != nil {
		utils.HandleError(c, validationErr)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(h.uploadConfig.Path, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(h.uploadConfig.Path, 0755); err != nil {
		utils.ErrorResponseJSON(c, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to create upload directory", nil)
		return
	}

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.ErrorResponseJSON(c, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to save file", nil)
		return
	}

	// Get user ID
	userID := middlewares.GetUserID(c)
	uploaderID, _ := uuid.Parse(userID)

	// Create media record
	media := &models.Media{
		Filename:         filename,
		OriginalFilename: file.Filename,
		FilePath:         filePath,
		FileSize:         file.Size,
		MimeType:         file.Header.Get("Content-Type"),
		UploadedBy:       uploaderID,
		AltText:          c.PostForm("alt_text"),
	}

	response, err := h.mediaService.UploadMedia(media)
	if err != nil {
		// Clean up file on error
		os.Remove(filePath)
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "File uploaded successfully", response)
}

// GetMedia returns a single media by ID
// @Summary Get media by ID
// @Description Get a single media file info by its ID
// @Tags media
// @Produce json
// @Param id path string true "Media ID (UUID)"
// @Success 200 {object} utils.Response{data=dto.MediaResponse} "Media retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 404 {object} utils.Response "Media not found"
// @Router /media/{id} [get]
func (h *MediaHandler) GetMedia(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Media ID is required", nil)
		return
	}

	media, err := h.mediaService.GetMedia(id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Media retrieved successfully", media)
}

// GetAllMedia returns all media
// @Summary List all media
// @Description Get a paginated list of all media files (admin only)
// @Tags media
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.MediaResponse} "Media retrieved successfully"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires admin role"
// @Router /media [get]
func (h *MediaHandler) GetAllMedia(c *gin.Context) {
	var query dto.MediaListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	media, total, err := h.mediaService.GetAllMedia(&query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Media retrieved successfully", media, meta)
}

// DeleteMedia deletes a media
// @Summary Delete media
// @Description Delete a media file (owner or admin only)
// @Tags media
// @Produce json
// @Security BearerAuth
// @Param id path string true "Media ID (UUID)"
// @Success 200 {object} utils.Response "Media deleted successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - not the owner"
// @Failure 404 {object} utils.Response "Media not found"
// @Router /media/{id} [delete]
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Media ID is required", nil)
		return
	}

	userID := middlewares.GetUserID(c)
	isAdmin := middlewares.IsAdmin(c)

	if err := h.mediaService.DeleteMedia(id, userID, isAdmin); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Media deleted successfully", nil)
}
