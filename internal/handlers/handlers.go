package handlers

import (
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/services"
	"github.com/yandas/backend/internal/websocket"
	"gorm.io/gorm"
)

// Handlers holds all handler instances
type Handlers struct {
	Auth         *AuthHandler
	User         *UserHandler
	Category     *CategoryHandler
	Yandas       *YandasHandler
	Order        *OrderHandler
	Chat         *ChatHandler
	Call         *CallHandler
	Subscription *SubscriptionHandler
	Notification *NotificationHandler
	Admin        *AdminHandler
	Legal        *LegalHandler
	Favorite     *FavoriteHandler
	Support      *SupportHandler
	Search       *SearchHandler
}

// NewHandlers creates all handlers
func NewHandlers(svcs *services.Services, cfg *config.Config, wsHub *websocket.Hub, db *gorm.DB) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(svcs),
		User:         NewUserHandler(svcs),
		Category:     NewCategoryHandler(svcs),
		Yandas:       NewYandasHandler(svcs),
		Order:        NewOrderHandler(svcs),
		Chat:         NewChatHandler(svcs, wsHub),
		Call:         NewCallHandler(svcs, wsHub, cfg, db),
		Subscription: NewSubscriptionHandler(svcs),
		Notification: NewNotificationHandler(svcs),
		Admin:        NewAdminHandler(svcs),
		Legal:        NewLegalHandler(cfg),
		Favorite:     NewFavoriteHandler(svcs),
		Support:      NewSupportHandler(svcs),
		Search:       NewSearchHandler(svcs),
	}
}

// Response helpers
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func SuccessResponse(data interface{}) Response {
	return Response{Success: true, Data: data}
}

func SuccessResponseWithMeta(data interface{}, meta *Meta) Response {
	return Response{Success: true, Data: data, Meta: meta}
}

func ErrorResponse(err string) Response {
	return Response{Success: false, Error: err}
}

func PaginationMeta(page, limit int, total int64) *Meta {
	totalPages := total / int64(limit)
	if total%int64(limit) > 0 {
		totalPages++
	}
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
