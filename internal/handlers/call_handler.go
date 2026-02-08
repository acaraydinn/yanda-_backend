package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/services"
	"github.com/yandas/backend/internal/websocket"
	"github.com/yandas/backend/pkg/agora"
	"gorm.io/gorm"
)

type CallHandler struct {
	svcs  *services.Services
	wsHub *websocket.Hub
	cfg   *config.Config
	db    *gorm.DB
}

func NewCallHandler(svcs *services.Services, wsHub *websocket.Hub, cfg *config.Config, db *gorm.DB) *CallHandler {
	return &CallHandler{svcs: svcs, wsHub: wsHub, cfg: cfg, db: db}
}

// InitiateCall starts a new call
func (h *CallHandler) InitiateCall(c *gin.Context) {
	callerID := getUserID(c)
	log.Printf("[CALL] InitiateCall: callerID=%s", callerID.String())

	var input struct {
		ReceiverID string `json:"receiver_id" binding:"required"`
		CallType   string `json:"call_type" binding:"required"` // "audio" or "video"
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("[CALL] InitiateCall: bind error: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse("receiver_id and call_type are required"))
		return
	}

	log.Printf("[CALL] InitiateCall: receiverID=%s, callType=%s", input.ReceiverID, input.CallType)

	receiverID, err := uuid.Parse(input.ReceiverID)
	if err != nil {
		log.Printf("[CALL] InitiateCall: invalid receiver_id: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse("invalid receiver_id"))
		return
	}

	if input.CallType != "audio" && input.CallType != "video" {
		c.JSON(http.StatusBadRequest, ErrorResponse("call_type must be 'audio' or 'video'"))
		return
	}

	// Generate unique channel name
	channelName := fmt.Sprintf("call_%s_%d", uuid.New().String()[:8], time.Now().Unix())

	// Generate Agora token for caller (uid = 1)
	token, err := agora.GenerateRTCToken(
		h.cfg.AgoraAppID,
		h.cfg.AgoraAppCertificate,
		channelName,
		1,    // caller UID
		3600, // 1 hour expiry
	)
	if err != nil {
		log.Printf("[CALL] InitiateCall: token generation error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to generate call token"))
		return
	}

	// Create call log
	callLog := &models.CallLog{
		ID:        uuid.New(),
		CallerID:  callerID,
		CalleeID:  receiverID,
		CallType:  input.CallType,
		Status:    "ringing",
		ChannelID: &channelName,
	}

	if err := h.db.Create(callLog).Error; err != nil {
		log.Printf("[CALL] InitiateCall: db create error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to create call"))
		return
	}

	// Get caller info for the notification
	var caller models.User
	h.db.First(&caller, "id = ?", callerID)
	log.Printf("[CALL] InitiateCall: caller=%s, callerName=%s", callerID.String(), caller.FullName)

	// Notify receiver via WebSocket
	log.Printf("[CALL] InitiateCall: Broadcasting incoming_call to receiverID=%s", receiverID.String())
	h.wsHub.BroadcastToUser(receiverID.String(), "incoming_call", map[string]interface{}{
		"call_id":       callLog.ID.String(),
		"caller_id":     callerID.String(),
		"caller_name":   caller.FullName,
		"caller_avatar": caller.AvatarURL,
		"call_type":     input.CallType,
		"channel_name":  channelName,
	})
	log.Printf("[CALL] InitiateCall: incoming_call broadcast DONE")

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"call_id":      callLog.ID.String(),
		"channel_name": channelName,
		"token":        token,
		"uid":          1,
		"app_id":       h.cfg.AgoraAppID,
	}))
}

// AnswerCall accepts an incoming call
func (h *CallHandler) AnswerCall(c *gin.Context) {
	callID, _ := uuid.Parse(c.Param("id"))
	userID := getUserID(c)

	var callLog models.CallLog
	if err := h.db.First(&callLog, "id = ? AND callee_id = ?", callID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("call not found"))
		return
	}

	if callLog.Status != "ringing" {
		c.JSON(http.StatusBadRequest, ErrorResponse("call is no longer ringing"))
		return
	}

	// Generate token for callee (uid = 2)
	token, err := agora.GenerateRTCToken(
		h.cfg.AgoraAppID,
		h.cfg.AgoraAppCertificate,
		*callLog.ChannelID,
		2, // callee UID
		3600,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to generate token"))
		return
	}

	// Update call status
	now := time.Now()
	h.db.Model(&callLog).Updates(map[string]interface{}{
		"status":      "answered",
		"answered_at": now,
	})

	// Notify caller that call was answered
	h.wsHub.BroadcastToUser(callLog.CallerID.String(), "call_answered", map[string]interface{}{
		"call_id": callLog.ID.String(),
	})

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"call_id":      callLog.ID.String(),
		"channel_name": *callLog.ChannelID,
		"token":        token,
		"uid":          2,
		"app_id":       h.cfg.AgoraAppID,
	}))
}

// RejectCall declines an incoming call
func (h *CallHandler) RejectCall(c *gin.Context) {
	callID, _ := uuid.Parse(c.Param("id"))
	userID := getUserID(c)

	var callLog models.CallLog
	if err := h.db.First(&callLog, "id = ? AND callee_id = ?", callID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("call not found"))
		return
	}

	now := time.Now()
	h.db.Model(&callLog).Updates(map[string]interface{}{
		"status":   "declined",
		"ended_at": now,
	})

	// Notify caller
	h.wsHub.BroadcastToUser(callLog.CallerID.String(), "call_rejected", map[string]interface{}{
		"call_id": callLog.ID.String(),
	})

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Call rejected"}))
}

// EndCall ends an active call
func (h *CallHandler) EndCall(c *gin.Context) {
	callID, _ := uuid.Parse(c.Param("id"))
	userID := getUserID(c)

	var callLog models.CallLog
	if err := h.db.First(&callLog, "id = ? AND (caller_id = ? OR callee_id = ?)", callID, userID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("call not found"))
		return
	}

	now := time.Now()
	duration := 0
	if callLog.AnsweredAt != nil {
		duration = int(now.Sub(*callLog.AnsweredAt).Seconds())
	}

	h.db.Model(&callLog).Updates(map[string]interface{}{
		"status":   "ended",
		"ended_at": now,
		"duration": duration,
	})

	// Notify the other party
	otherUserID := callLog.CallerID.String()
	if callLog.CallerID == userID {
		otherUserID = callLog.CalleeID.String()
	}

	h.wsHub.BroadcastToUser(otherUserID, "call_ended", map[string]interface{}{
		"call_id":  callLog.ID.String(),
		"duration": duration,
	})

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Call ended", "duration": duration}))
}
