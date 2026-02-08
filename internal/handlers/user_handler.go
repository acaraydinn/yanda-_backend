package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/services"
)

type UserHandler struct {
	svcs *services.Services
}

func NewUserHandler(svcs *services.Services) *UserHandler {
	return &UserHandler{svcs: svcs}
}

func getUserID(c *gin.Context) uuid.UUID {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	return userID
}

func getPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	return page, limit
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	user, err := h.svcs.User.GetProfile(getUserID(c))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(user))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var input services.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	user, err := h.svcs.User.UpdateProfile(getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(user))
}

func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("avatar required"))
		return
	}
	userID := getUserID(c)
	dst := "./uploads/avatars/" + userID.String() + "_" + file.Filename
	c.SaveUploadedFile(file, dst)
	h.svcs.User.UpdateAvatar(userID, dst)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"avatar_url": dst}))
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var input services.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	if err := h.svcs.User.ChangePassword(getUserID(c), &input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Password changed"}))
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	if err := h.svcs.User.DeleteAccount(getUserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Account deleted"}))
}

func (h *UserHandler) RegisterDeviceToken(c *gin.Context) {
	var input struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	h.svcs.User.RegisterDeviceToken(getUserID(c), input.Token, input.Platform)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Token registered"}))
}
