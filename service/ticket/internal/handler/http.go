package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"ticket-tix/service/ticket/internal/model"
	"ticket-tix/service/ticket/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	service *service.TicketService
}

func NewTicketHandler(svc *service.TicketService) *TicketHandler {
	return &TicketHandler{service: svc}
}

func (h *TicketHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/events", h.CreateEvent)
	router.POST("/events/:id/images", h.UploadImage)
	router.DELETE("/events/:id/images/:imageID", h.DeleteImage)
}

type createEventRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	Location    string `form:"location" binding:"required"`
	StartTime   string `form:"start_time" binding:"required"`
	EndTime     string `form:"end_time" binding:"required"`
}

func (h *TicketHandler) CreateEvent(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time, use RFC3339"})
		return
	}

	files, err := constructFilesFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceReq := model.InsertTicketRequest{
		Event: model.EventData{
			Name:        req.Name,
			Description: req.Description,
			Location:    req.Location,
			StartTime:   startTime,
			EndTime:     endTime,
		},
		Files: files,
	}

	result, err := h.service.CreateEvent(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *TicketHandler) UploadImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	files, err := constructFilesFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UploadImg(c.Request.Context(), int32(id), files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "images uploaded"})
}

func (h *TicketHandler) DeleteImage(c *gin.Context) {
	key := c.Param("imageID")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "imageID is required"})
		return
	}

	if err := h.service.DeleteImg(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

func constructFilesFromRequest(c *gin.Context) ([]model.FileData, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("failed to read multipart form: %w", err)
	}

	fileHeaders := form.File["images"]
	if len(fileHeaders) == 0 {
		return nil, fmt.Errorf("at least one image is required")
	}

	var files []model.FileData
	for i, header := range fileHeaders {
		file, err := header.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open image %s: %w", header.Filename, err)
		}

		files = append(files, model.FileData{
			Filename:     header.Filename,
			Size:         header.Size,
			ContentType:  header.Header.Get("Content-Type"),
			Reader:       file,
			IsPrimary:    i == 0,
			DisplayOrder: i,
		})
	}

	return files, nil
}
