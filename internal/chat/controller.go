package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	ChatService *ChatService
}

func NewChatController(cs *ChatService) *ChatController {
	return &ChatController{ChatService: cs}
}

func (cc *ChatController) Chat(c *gin.Context) {
	question := c.PostForm("question")
	filename := c.PostForm("filename")
	audioFile, _ := c.FormFile("audio")

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	answer, err := cc.ChatService.Chat(question, audioFile, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
