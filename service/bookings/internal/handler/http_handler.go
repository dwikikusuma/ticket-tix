package handler

import (
	"ticket-tix/common/pkg/middleware"
	"ticket-tix/service/bookings/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	secretKey = "sudo-secret-key"
)

type Handler struct {
	service     model.BookingService
	redisClient *redis.Client
}

func NewHandler(service model.BookingService, rdsClient *redis.Client) *Handler {
	return &Handler{
		service:     service,
		redisClient: rdsClient,
	}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := r.Group("/bookings")
	auth.Use(middleware.AuthMiddleware(secretKey, h.redisClient))
	auth.POST("/create", h.CreateBooking)
}

func (h *Handler) CreateBooking(c *gin.Context) {
	var req model.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt32("userID")

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
