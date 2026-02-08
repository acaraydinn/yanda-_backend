package repository

import (
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// SubscriptionRepository handles subscription operations
type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *models.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *SubscriptionRepository) GetByUserID(userID uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").First(&sub).Error
	return &sub, err
}

func (r *SubscriptionRepository) GetByProviderID(providerID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.First(&sub, "provider_subscription_id = ?", providerID).Error
	return &sub, err
}

func (r *SubscriptionRepository) Update(sub *models.Subscription) error {
	return r.db.Save(sub).Error
}

func (r *SubscriptionRepository) Cancel(id uuid.UUID) error {
	return r.db.Model(&models.Subscription{}).
		Where("id = ?", id).
		Update("status", "cancelled").Error
}

// DeviceTokenRepository handles device token operations
type DeviceTokenRepository struct {
	db *gorm.DB
}

func NewDeviceTokenRepository(db *gorm.DB) *DeviceTokenRepository {
	return &DeviceTokenRepository{db: db}
}

func (r *DeviceTokenRepository) Create(token *models.DeviceToken) error {
	// First try to find existing token
	var existing models.DeviceToken
	err := r.db.Where("user_id = ? AND token = ?", token.UserID, token.Token).First(&existing).Error
	if err == nil {
		// Token exists, update it
		existing.IsActive = true
		return r.db.Save(&existing).Error
	}
	return r.db.Create(token).Error
}

func (r *DeviceTokenRepository) GetByUserID(userID uuid.UUID) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&tokens).Error
	return tokens, err
}

func (r *DeviceTokenRepository) Deactivate(token string) error {
	return r.db.Model(&models.DeviceToken{}).
		Where("token = ?", token).
		Update("is_active", false).Error
}

func (r *DeviceTokenRepository) DeactivateAllForUser(userID uuid.UUID) error {
	return r.db.Model(&models.DeviceToken{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

// AuditLogRepository handles audit log operations
type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditLogRepository) List(page, limit int, adminID *uuid.UUID, action string) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{})
	if adminID != nil {
		query = query.Where("admin_id = ?", *adminID)
	}
	if action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Admin").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, total, err
}

// NotificationRepository handles notification operations
type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notif *models.Notification) error {
	return r.db.Create(notif).Error
}

func (r *NotificationRepository) ListByUser(userID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	var notifs []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&notifs).Error

	return notifs, total, err
}

func (r *NotificationRepository) MarkAsRead(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

func (r *NotificationRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
