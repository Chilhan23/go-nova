package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	chatService Service
}

func NewHandler(chatService Service) *Handler {
	return &Handler{chatService: chatService}
}

func (h *Handler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	res, err := h.chatService.SendMessage(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) UploadAttachment(c *gin.Context) {
	ticketIDStr := c.PostForm("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil || ticketID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid ticket_id"})
		return
	}

	file, err := c.FormFile("attachment")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "No attachment file uploaded"})
		return
	}

	caption := c.PostForm("caption")
	res, err := h.chatService.UploadAttachment(c.Request.Context(), ticketID, file, caption)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) GetHistory(c *gin.Context) {
	ticketIDStr := c.Param("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid ticket_id"})
		return
	}

	messages, err := h.chatService.GetHistory(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   messages,
	})
}
