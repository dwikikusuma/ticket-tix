package handler

import "github.com/gin-gonic/gin"

type TicketHandler struct {
}

func NewTicketHandler() *TicketHandler {
	return &TicketHandler{}
}

func (h *TicketHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", func(context *gin.Context) {
		context.JSON(200, gin.H{"status": "ok"})
	})
}
