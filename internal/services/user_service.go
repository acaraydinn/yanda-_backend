package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user operations
type UserService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

// NewUserService creates a new user service
func NewUserService(repos *repository.Repositories, cfg *config.Config) *UserService {
	return &UserService{repos: repos, cfg: cfg}
}

// GetProfile returns user profile
func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.repos.User.GetByID(userID)
}

// UpdateProfileInput represents profile update data
type UpdateProfileInput struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(userID uuid.UUID, input *UpdateProfileInput) (*models.User, error) {
	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if input.FullName != "" {
		user.FullName = input.FullName
	}

	if input.Phone != "" {
		// Check if phone is already taken
		if s.repos.User.ExistsByPhone(input.Phone) {
			existing, _ := s.repos.User.GetByPhone(input.Phone)
			if existing != nil && existing.ID != userID {
				return nil, errors.New("phone number already in use")
			}
		}
		user.Phone = &input.Phone
	}

	if err := s.repos.User.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateAvatar updates user avatar
func (s *UserService) UpdateAvatar(userID uuid.UUID, avatarURL string) error {
	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return err
	}

	user.AvatarURL = &avatarURL
	return s.repos.User.Update(user)
}

// ChangePasswordInput represents password change data
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(userID uuid.UUID, input *ChangePasswordInput) error {
	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.repos.User.Update(user)
}

// DeleteAccount deletes user account (GDPR compliant)
func (s *UserService) DeleteAccount(userID uuid.UUID) error {
	// Deactivate all device tokens
	s.repos.DeviceToken.DeactivateAllForUser(userID)

	// Hard delete user data
	return s.repos.User.HardDelete(userID)
}

// RegisterDeviceToken registers a device token for push notifications
func (s *UserService) RegisterDeviceToken(userID uuid.UUID, token, platform string) error {
	deviceToken := &models.DeviceToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
		IsActive: true,
	}
	return s.repos.DeviceToken.Create(deviceToken)
}
