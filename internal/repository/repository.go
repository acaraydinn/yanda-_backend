package repository

import (
	"gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
	User          *UserRepository
	YandasProfile *YandasProfileRepository
	Category      *CategoryRepository
	Service       *ServiceRepository
	Order         *OrderRepository
	Review        *ReviewRepository
	Conversation  *ConversationRepository
	Message       *MessageRepository
	Subscription  *SubscriptionRepository
	DeviceToken   *DeviceTokenRepository
	AuditLog      *AuditLogRepository
	Notification  *NotificationRepository
	Support       *SupportRepository
	Favorite      *FavoriteRepository
}

// NewRepositories creates all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:          NewUserRepository(db),
		YandasProfile: NewYandasProfileRepository(db),
		Category:      NewCategoryRepository(db),
		Service:       NewServiceRepository(db),
		Order:         NewOrderRepository(db),
		Review:        NewReviewRepository(db),
		Conversation:  NewConversationRepository(db),
		Message:       NewMessageRepository(db),
		Subscription:  NewSubscriptionRepository(db),
		DeviceToken:   NewDeviceTokenRepository(db),
		AuditLog:      NewAuditLogRepository(db),
		Notification:  NewNotificationRepository(db),
		Support:       NewSupportRepository(db),
		Favorite:      NewFavoriteRepository(db),
	}
}
