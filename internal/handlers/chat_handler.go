package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/services"
	"github.com/yandas/backend/internal/websocket"
)

type ChatHandler struct {
	svcs  *services.Services
	wsHub *websocket.Hub
}

func NewChatHandler(svcs *services.Services, wsHub *websocket.Hub) *ChatHandler {
	return &ChatHandler{svcs: svcs, wsHub: wsHub}
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	page, limit := getPagination(c)
	convs, total, _ := h.svcs.Chat.GetConversations(getUserID(c), page, limit)
	c.JSON(http.StatusOK, SuccessResponseWithMeta(convs, PaginationMeta(page, limit, total)))
}

func (h *ChatHandler) GetConversation(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	conv, err := h.svcs.Chat.GetConversation(getUserID(c), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(conv))
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	page, limit := getPagination(c)
	msgs, total, _ := h.svcs.Chat.GetMessages(getUserID(c), id, page, limit)
	c.JSON(http.StatusOK, SuccessResponseWithMeta(msgs, PaginationMeta(page, limit, total)))
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input services.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	msg, err := h.svcs.Chat.SendMessage(getUserID(c), id, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	// Broadcast via WebSocket
	h.wsHub.BroadcastToConversation(id.String(), msg)
	c.JSON(http.StatusCreated, SuccessResponse(msg))
}

func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Chat.MarkAsRead(getUserID(c), id)

	// Broadcast read receipt via WebSocket
	h.wsHub.BroadcastToConversation(id.String(), map[string]interface{}{
		"type":            "read",
		"conversation_id": id.String(),
		"reader_id":       getUserID(c).String(),
	})

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Marked as read"}))
}

// SendImageMessage handles image/file upload in chat
func (h *ChatHandler) SendImageMessage(c *gin.Context) {
	convID, _ := uuid.Parse(c.Param("id"))
	userID := getUserID(c)

	// Save the uploaded file using the existing helper
	filePath, err := saveUploadedFile(c, "image", userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("image file required"))
		return
	}

	// Create message with image type
	input := &services.SendMessageInput{
		Content:     filePath,
		MessageType: "image",
	}

	msg, err := h.svcs.Chat.SendMessage(userID, convID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	// Broadcast via WebSocket
	h.wsHub.BroadcastToConversation(convID.String(), msg)
	c.JSON(http.StatusCreated, SuccessResponse(msg))
}

// StartConversation starts a new chat conversation with a yandaş
func (h *ChatHandler) StartConversation(c *gin.Context) {
	var input struct {
		YandasUserID string `json:"yandas_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("yandas_user_id is required"))
		return
	}

	yandasUserID, err := uuid.Parse(input.YandasUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("invalid yandas_user_id"))
		return
	}

	conv, err := h.svcs.Chat.StartConversation(getUserID(c), yandasUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(conv))
}

type SubscriptionHandler struct {
	svcs *services.Services
}

func NewSubscriptionHandler(svcs *services.Services) *SubscriptionHandler {
	return &SubscriptionHandler{svcs: svcs}
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	sub, err := h.svcs.Subscription.Get(getUserID(c))
	if err != nil {
		c.JSON(http.StatusOK, SuccessResponse(nil))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(sub))
}

func (h *SubscriptionHandler) Verify(c *gin.Context) {
	var input services.VerifyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	sub, err := h.svcs.Subscription.Verify(getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(sub))
}

func (h *SubscriptionHandler) Webhook(c *gin.Context) {
	body, _ := c.GetRawData()
	h.svcs.Subscription.HandleWebhook(body)
	c.JSON(http.StatusOK, gin.H{"received": true})
}

type NotificationHandler struct {
	svcs *services.Services
}

func NewNotificationHandler(svcs *services.Services) *NotificationHandler {
	return &NotificationHandler{svcs: svcs}
}

func (h *NotificationHandler) List(c *gin.Context) {
	page, limit := getPagination(c)
	notifs, total, _ := h.svcs.Notification.List(getUserID(c), page, limit)
	c.JSON(http.StatusOK, SuccessResponseWithMeta(notifs, PaginationMeta(page, limit, total)))
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Notification.MarkAsRead(id)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Marked"}))
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	h.svcs.Notification.MarkAllAsRead(getUserID(c))
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "All marked"}))
}
