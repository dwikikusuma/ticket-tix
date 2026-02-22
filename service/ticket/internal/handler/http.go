package handler

import (
	"net/http"
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
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/events", h.CreateEvent)
}

func (h *TicketHandler) CreateEvent(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, c.PostForm("start_time"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, c.PostForm("end_time"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time, use RFC3339"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form"})
		return
	}

	fileHeaders := form.File["images"]
	if len(fileHeaders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one image is required"})
		return
	}

	var files []model.FileData
	for i, header := range fileHeaders {
		file, err := header.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open image"})
			return
		}
		defer file.Close()

		files = append(files, model.FileData{
			Filename:     header.Filename,
			Size:         header.Size,
			ContentType:  header.Header.Get("Content-Type"),
			Reader:       file,
			IsPrimary:    i == 0, // first image is primary
			DisplayOrder: i,
		})
	}

	req := model.InsertTicketRequest{
		Event: model.EventData{
			Name:        c.PostForm("name"),
			Description: c.PostForm("description"),
			Location:    c.PostForm("location"),
			StartTime:   startTime,
			EndTime:     endTime,
		},
		Files: files,
	}

	result, err := h.service.CreateEvent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}
