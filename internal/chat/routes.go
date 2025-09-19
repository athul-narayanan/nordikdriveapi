package chat

import (
	"nordik-drive-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, chatService *ChatService) {
	chatController := &ChatController{ChatService: chatService}

	userGroup := r.Group("/api/chat")
	userGroup.Use(middlewares.AuthMiddleware())
	{
		userGroup.POST("", chatController.Chat)
	}

}
