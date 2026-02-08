package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
)

// SubscriptionService handles subscription operations
type SubscriptionService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func NewSubscriptionService(repos *repository.Repositories, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{repos: repos, cfg: cfg}
}

// Get returns user subscription
func (s *SubscriptionService) Get(userID uuid.UUID) (*models.Subscription, error) {
	return s.repos.Subscription.GetByUserID(userID)
}

// VerifyInput represents subscription verification data from RevenueCat
type VerifyInput struct {
	ReceiptData  string `json:"receipt_data" binding:"required"`
	ProductID    string `json:"product_id" binding:"required"`
	Platform     string `json:"platform" binding:"required"` // ios, android
	IsRestore    bool   `json:"is_restore"`
}

// Verify verifies and creates/updates subscription
func (s *SubscriptionService) Verify(userID uuid.UUID, input *VerifyInput) (*models.Subscription, error) {
	// TODO: Verify with RevenueCat API
	// For now, just create the subscription

	planType := "monthly"
	if input.ProductID == "yandas_pro_yearly" {
		planType = "yearly"
	}

	now := time.Now()
	var periodEnd time.Time
	if planType == "monthly" {
		periodEnd = now.AddDate(0, 1, 0)
	} else {
		periodEnd = now.AddDate(1, 0, 0)
	}

	// Check if subscription exists
	existing, _ := s.repos.Subscription.GetByUserID(userID)
	if existing != nil {
		existing.PlanType = planType
		existing.CurrentPeriodStart = &now
		existing.CurrentPeriodEnd = &periodEnd
		existing.Status = "active"
		if err := s.repos.Subscription.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	sub := &models.Subscription{
		UserID:             userID,
		PlanType:           planType,
		Status:             "active",
		Provider:           "revenuecat",
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   &periodEnd,
	}

	if err := s.repos.Subscription.Create(sub); err != nil {
		return nil, err
	}

	// Update user role to yandas if not already
	user, _ := s.repos.User.GetByID(userID)
	if user != nil && user.Role == "customer" {
		// Only if they have an approved yandas profile
		profile, _ := s.repos.YandasProfile.GetByUserID(userID)
		if profile != nil && profile.ApprovalStatus == "approved" {
			user.Role = "yandas"
			s.repos.User.Update(user)
		}
	}

	return sub, nil
}

// WebhookPayload represents RevenueCat webhook payload
type WebhookPayload struct {
	Event struct {
		Type                  string `json:"type"`
		AppUserID             string `json:"app_user_id"`
		ProductID             string `json:"product_id"`
		OriginalTransactionID string `json:"original_transaction_id"`
		ExpirationAtMs        int64  `json:"expiration_at_ms"`
	} `json:"event"`
}

// HandleWebhook handles RevenueCat webhook
func (s *SubscriptionService) HandleWebhook(payload []byte) error {
	var webhook WebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return err
	}

	userID, err := uuid.Parse(webhook.Event.AppUserID)
	if err != nil {
		return err
	}

	sub, err := s.repos.Subscription.GetByUserID(userID)
	if err != nil {
		return nil // No subscription to update
	}

	switch webhook.Event.Type {
	case "INITIAL_PURCHASE", "RENEWAL":
		expiration := time.UnixMilli(webhook.Event.ExpirationAtMs)
		sub.Status = "active"
		sub.CurrentPeriodEnd = &expiration
	case "CANCELLATION":
		now := time.Now()
		sub.Status = "cancelled"
		sub.CancelledAt = &now
	case "EXPIRATION":
		sub.Status = "expired"
	}

	return s.repos.Subscription.Update(sub)
}

// NotificationService handles notification operations
type NotificationService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func NewNotificationService(repos *repository.Repositories, cfg *config.Config) *NotificationService {
	return &NotificationService{repos: repos, cfg: cfg}
}

func (s *NotificationService) List(userID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	return s.repos.Notification.ListByUser(userID, page, limit)
}

func (s *NotificationService) MarkAsRead(notificationID uuid.UUID) error {
	return s.repos.Notification.MarkAsRead(notificationID)
}

func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.repos.Notification.MarkAllAsRead(userID)
}

// Send creates a notification and sends push
func (s *NotificationService) Send(userID uuid.UUID, title, body, notifType string, data map[string]interface{}) error {
	// Create in-app notification
	var dataStr *string
	if data != nil {
		dataBytes, _ := json.Marshal(data)
		str := string(dataBytes)
		dataStr = &str
	}

	notif := &models.Notification{
		UserID: userID,
		Title:  title,
		Body:   body,
		Type:   notifType,
		Data:   dataStr,
	}

	if err := s.repos.Notification.Create(notif); err != nil {
		return err
	}

	// Send push notification
	go s.sendPush(userID, title, body, data)

	return nil
}

func (s *NotificationService) sendPush(userID uuid.UUID, title, body string, data map[string]interface{}) {
	tokens, err := s.repos.DeviceToken.GetByUserID(userID)
	if err != nil {
		return
	}

	// TODO: Send via FCM
	_ = tokens
}
