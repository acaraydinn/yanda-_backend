package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	twilio "github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
	"github.com/yandas/backend/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
	ErrUserNotVerified    = errors.New("user not verified")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService handles authentication
type AuthService struct {
	repos    *repository.Repositories
	cfg      *config.Config
	redis    *redis.Client
	emailSvc *EmailService
}

// NewAuthService creates a new auth service
func NewAuthService(repos *repository.Repositories, cfg *config.Config, redis *redis.Client, emailSvc *EmailService) *AuthService {
	return &AuthService{repos: repos, cfg: cfg, redis: redis, emailSvc: emailSvc}
}

// RegisterInput represents registration data
type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Platform string `json:"platform"`
}

// Register creates a new user account
func (s *AuthService) Register(input *RegisterInput) (*models.User, *auth.TokenPair, error) {
	// Check if user exists
	if s.repos.User.ExistsByEmail(input.Email) {
		return nil, nil, ErrUserExists
	}

	if input.Phone != "" && s.repos.User.ExistsByPhone(input.Phone) {
		return nil, nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        &input.Email,
		PasswordHash: string(hashedPassword),
		FullName:     input.FullName,
		Role:         "customer",
		IsVerified:   false,
		IsActive:     true,
	}

	if input.Phone != "" {
		user.Phone = &input.Phone
	}

	if err := s.repos.User.Create(user); err != nil {
		return nil, nil, err
	}

	// Generate tokens
	tokens, err := auth.GenerateTokenPair(
		user.ID.String(),
		*user.Email,
		user.Role,
		input.Platform,
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, nil, err
	}

	// Send email OTP
	go func() {
		if err := s.SendEmailOTP(input.Email, input.FullName); err != nil {
			log.Printf("Email OTP gönderme hatası: %v\n", err)
		}
	}()

	// Send SMS OTP if phone provided
	if input.Phone != "" {
		go func() {
			if err := s.SendOTP(input.Phone); err != nil {
				log.Printf("SMS OTP gönderme hatası: %v\n", err)
			}
		}()
	}

	return user, tokens, nil
}

// LoginInput represents login data
type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Platform string `json:"platform"`
}

// Login authenticates a user
func (s *AuthService) Login(input *LoginInput) (*models.User, *auth.TokenPair, error) {
	user, err := s.repos.User.GetByEmail(input.Email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	tokens, err := auth.GenerateTokenPair(
		user.ID.String(),
		email,
		user.Role,
		input.Platform,
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// RefreshToken generates new token pair from refresh token
func (s *AuthService) RefreshToken(refreshToken, platform string) (*auth.TokenPair, error) {
	claims, err := auth.ValidateToken(refreshToken, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	// Verify user still exists and is active
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	return auth.GenerateTokenPair(
		user.ID.String(),
		email,
		user.Role,
		platform,
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
}

// SendOTP sends OTP to phone number via Twilio Verify
func (s *AuthService) SendOTP(phone string) error {
	// Normalize phone to E.164 format for Twilio
	phone = normalizePhone(phone)

	// Eğer Twilio yapılandırılmamışsa fallback kullan
	if s.cfg.TwilioAccountSID == "" || s.cfg.TwilioVerifySID == "" {
		otp := generateOTP()
		if s.redis != nil {
			key := fmt.Sprintf("otp:%s", phone)
			ctx := context.Background()
			s.redis.Set(ctx, key, otp, 5*time.Minute)
		}
		log.Printf("[FALLBACK] OTP for %s: %s\n", phone, otp)
		return nil
	}

	// Twilio Verify ile SMS gönder
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.cfg.TwilioAccountSID,
		Password: s.cfg.TwilioAuthToken,
	})

	channel := "sms"
	params := &verify.CreateVerificationParams{}
	params.SetTo(phone)
	params.SetChannel(channel)

	_, err := client.VerifyV2.CreateVerification(s.cfg.TwilioVerifySID, params)
	if err != nil {
		log.Printf("Twilio SMS gönderme hatası: %v\n", err)
		return fmt.Errorf("SMS gönderilemedi: %w", err)
	}

	log.Printf("✅ OTP SMS gönderildi: %s\n", phone)
	return nil
}

// normalizePhone converts Turkish phone numbers to E.164 format
func normalizePhone(phone string) string {
	// Remove spaces, dashes, parentheses
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	if strings.HasPrefix(phone, "+") {
		return phone // Already international format
	}
	if strings.HasPrefix(phone, "0") {
		return "+90" + phone[1:] // 05xx -> +905xx
	}
	if strings.HasPrefix(phone, "5") {
		return "+90" + phone // 5xx -> +905xx
	}
	return "+90" + phone
}

// VerifyOTP verifies OTP code via Twilio Verify
func (s *AuthService) VerifyOTP(phone, otp string) error {
	originalPhone := phone
	phone = normalizePhone(phone)
	log.Printf("📞 VerifyOTP: giriş=%s, normalize=%s, otp=%s\n", originalPhone, phone, otp)

	// Eğer Twilio yapılandırılmamışsa Redis fallback kullan
	if s.cfg.TwilioAccountSID == "" || s.cfg.TwilioVerifySID == "" {
		if s.redis == nil {
			return ErrInvalidOTP
		}
		key := fmt.Sprintf("otp:%s", phone)
		ctx := context.Background()
		storedOTP, err := s.redis.Get(ctx, key).Result()
		if err != nil || storedOTP != otp {
			return ErrInvalidOTP
		}
		s.redis.Del(ctx, key)
	} else {
		// Twilio Verify ile doğrula
		client := twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: s.cfg.TwilioAccountSID,
			Password: s.cfg.TwilioAuthToken,
		})

		params := &verify.CreateVerificationCheckParams{}
		params.SetTo(phone)
		params.SetCode(otp)

		resp, err := client.VerifyV2.CreateVerificationCheck(s.cfg.TwilioVerifySID, params)
		if err != nil {
			log.Printf("Twilio doğrulama hatası: %v\n", err)
			return ErrInvalidOTP
		}

		statusStr := "nil"
		if resp.Status != nil {
			statusStr = *resp.Status
		}
		log.Printf("📞 Twilio OTP sonuç: status=%s, phone=%s\n", statusStr, phone)

		if resp.Status == nil || *resp.Status != "approved" {
			log.Printf("❌ Twilio OTP durumu: %s (beklenen: approved)\n", statusStr)
			return ErrInvalidOTP
		}

		log.Printf("✅ Telefon OTP doğrulandı: %s\n", phone)
	}

	return nil
}

// ForgotPassword initiates password reset
func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.repos.User.GetByEmail(email)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Generate reset token
	resetToken := uuid.New().String()

	if s.redis != nil {
		key := fmt.Sprintf("reset:%s", resetToken)
		ctx := context.Background()
		s.redis.Set(ctx, key, user.ID.String(), 1*time.Hour)
	}

	// TODO: Send reset email
	fmt.Printf("Reset token for %s: %s\n", email, resetToken)

	return nil
}

// ResetPassword resets password with token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	if s.redis == nil {
		return errors.New("service unavailable")
	}

	key := fmt.Sprintf("reset:%s", token)
	ctx := context.Background()

	userIDStr, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.repos.User.Update(user); err != nil {
		return err
	}

	// Delete reset token
	s.redis.Del(ctx, key)

	return nil
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// SendEmailOTP generates and sends email OTP
func (s *AuthService) SendEmailOTP(email, userName string) error {
	otp := generateOTP()

	// Store OTP in Redis
	if s.redis != nil {
		key := fmt.Sprintf("email_otp:%s", email)
		ctx := context.Background()
		s.redis.Set(ctx, key, otp, 5*time.Minute)
	}

	// Send email
	if s.emailSvc != nil {
		if err := s.emailSvc.SendOTPEmail(email, otp, userName); err != nil {
			log.Printf("Email OTP gönderme hatası: %v\n", err)
			return err
		}
	} else {
		log.Printf("[FALLBACK] Email OTP for %s: %s\n", email, otp)
	}

	return nil
}

// VerifyEmailOTP verifies email OTP code
func (s *AuthService) VerifyEmailOTP(email, otp string) error {
	if s.redis == nil {
		return ErrInvalidOTP
	}

	key := fmt.Sprintf("email_otp:%s", email)
	ctx := context.Background()

	storedOTP, err := s.redis.Get(ctx, key).Result()
	if err != nil || storedOTP != otp {
		return ErrInvalidOTP
	}

	// Delete OTP after successful verification
	s.redis.Del(ctx, key)
	return nil
}

// ResendEmailOTP resends email OTP
func (s *AuthService) ResendEmailOTP(email string) error {
	user, err := s.repos.User.GetByEmail(email)
	if err != nil {
		return ErrUserNotFound
	}
	return s.SendEmailOTP(email, user.FullName)
}

// VerifyAccount verifies both email and phone OTP, marks user as verified
func (s *AuthService) VerifyAccount(email, emailOTP, phone, phoneOTP string) error {
	log.Printf("🔐 Hesap doğrulama başlatıldı: email=%s, phone=%s\n", email, phone)

	// Verify email OTP
	if err := s.VerifyEmailOTP(email, emailOTP); err != nil {
		log.Printf("❌ E-posta OTP doğrulama başarısız: email=%s, err=%v\n", email, err)
		return fmt.Errorf("e-posta doğrulama kodu hatalı veya süresi dolmuş")
	}
	log.Printf("✅ E-posta OTP doğrulandı: %s\n", email)

	// Verify phone OTP if phone provided
	if phone != "" && phoneOTP != "" {
		log.Printf("📱 Telefon OTP doğrulama başlatıldı: %s\n", phone)
		if err := s.VerifyOTP(phone, phoneOTP); err != nil {
			log.Printf("❌ Telefon OTP doğrulama başarısız: phone=%s, err=%v\n", phone, err)
			return fmt.Errorf("telefon doğrulama kodu hatalı veya süresi dolmuş")
		}
		log.Printf("✅ Telefon OTP doğrulandı: %s\n", phone)
	}

	// Mark user as verified
	user, err := s.repos.User.GetByEmail(email)
	if err != nil {
		return ErrUserNotFound
	}

	user.IsVerified = true
	if err := s.repos.User.Update(user); err != nil {
		return err
	}

	// Send welcome email
	if s.emailSvc != nil {
		go s.emailSvc.SendWelcomeEmail(email, user.FullName)
	}

	log.Printf("✅ Hesap doğrulandı: %s\n", email)
	return nil
}
