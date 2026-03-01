package handler

import (
	"net/http"
	"ticket-tix/service/bookings/internal/model"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service model.BookingService
}

func NewHandler(service model.BookingService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/bookings", h.CreateBooking)
}

func (h *Handler) CreateBooking(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.BookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if bookErr := h.service.CreateBooking(ctx, req.EventID, req.EventCat, req.SeatID); bookErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": bookErr.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Booking created successfully"})
}
