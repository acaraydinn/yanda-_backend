package services

import (
	"github.com/redis/go-redis/v9"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/repository"
)

// Services holds all service instances
type Services struct {
	Auth         *AuthService
	User         *UserService
	Yandas       *YandasService
	Category     *CategoryService
	Order        *OrderService
	Chat         *ChatService
	Subscription *SubscriptionService
	Notification *NotificationService
	Admin        *AdminService
	Favorite     *FavoriteService
	Support      *SupportService
	Email        *EmailService
}

// NewServices creates all services
func NewServices(repos *repository.Repositories, cfg *config.Config, redis *redis.Client) *Services {
	emailSvc := NewEmailService(cfg)

	return &Services{
		Auth:         NewAuthService(repos, cfg, redis, emailSvc),
		User:         NewUserService(repos, cfg),
		Yandas:       NewYandasService(repos, cfg),
		Category:     NewCategoryService(repos),
		Order:        NewOrderService(repos, cfg),
		Chat:         NewChatService(repos),
		Subscription: NewSubscriptionService(repos, cfg),
		Notification: NewNotificationService(repos, cfg),
		Admin:        NewAdminService(repos),
		Favorite:     NewFavoriteService(repos),
		Support:      NewSupportService(repos),
		Email:        emailSvc,
	}
}
