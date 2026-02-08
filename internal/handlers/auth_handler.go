package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yandas/backend/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	svcs *services.Services
}

func NewAuthHandler(svcs *services.Services) *AuthHandler {
	return &AuthHandler{svcs: svcs}
}

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body services.RegisterInput true "Registration data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	input.Platform = c.GetHeader("X-Platform")
	if input.Platform == "" {
		input.Platform = "unknown"
	}

	user, tokens, err := h.svcs.Auth.Register(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(gin.H{
		"user":               user,
		"tokens":             tokens,
		"needs_verification": true,
	}))
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body services.LoginInput true "Login data"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	input.Platform = c.GetHeader("X-Platform")
	if input.Platform == "" {
		input.Platform = "unknown"
	}

	user, tokens, err := h.svcs.Auth.Login(&input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"user":               user,
		"tokens":             tokens,
		"needs_verification": !user.IsVerified,
	}))
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Refresh token"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	platform := c.GetHeader("X-Platform")
	tokens, err := h.svcs.Auth.RefreshToken(input.RefreshToken, platform)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(tokens))
}

// ForgotPassword godoc
// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Email"
// @Success 200 {object} Response
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	// Always return success to prevent email enumeration
	h.svcs.Auth.ForgotPassword(input.Email)

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"message": "If the email exists, a reset link has been sent",
	}))
}

// ResetPassword godoc
// @Summary Reset password with token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Token and new password"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.svcs.Auth.ResetPassword(input.Token, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"message": "Password has been reset successfully",
	}))
}

// VerifyPhone godoc
// @Summary Verify phone with OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Phone and OTP"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /auth/verify-phone [post]
func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.svcs.Auth.VerifyOTP(input.Phone, input.OTP); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"message": "Phone verified successfully",
	}))
}

// ResendOTP godoc
// @Summary Resend OTP to phone
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Phone"
// @Success 200 {object} Response
// @Router /auth/resend-otp [post]
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	h.svcs.Auth.SendOTP(input.Phone)

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"message": "OTP sent",
	}))
}

// VerifyAccount godoc
// @Summary Verify account with email and phone OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Email OTP and Phone OTP"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /auth/verify-account [post]
func (h *AuthHandler) VerifyAccount(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		EmailOTP string `json:"email_otp" binding:"required"`
		Phone    string `json:"phone"`
		PhoneOTP string `json:"phone_otp"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.svcs.Auth.VerifyAccount(input.Email, input.EmailOTP, input.Phone, input.PhoneOTP); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"message":  "Hesabınız başarıyla doğrulandı!",
		"verified": true,
	}))
}

// ResendEmailOTP godoc
// @Summary Resend email OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Email"
// @Success 200 {object} Response
// @Router /auth/resend-email-otp [post]
func (h *AuthHandler) ResendEmailOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.svcs.Auth.ResendEmailOTP(input.Email); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"message": "Doğrulama kodu e-postanıza gönderildi",
	}))
}
