package handler

import (
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
	var req model.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// For now, hardcode userID = 1 here (TIX-015 adds JWT extraction)
	userID := int32(1)

	if err := h.service.CreateBooking(
		c.Request.Context(),
		userID,
		req.EventID,
		req.EventCat,
		req.SeatID,
		req.BookType,
		req.CategoryType,
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "booking created"})
}
