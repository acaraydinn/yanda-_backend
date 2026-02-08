package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
)

// OrderService handles order operations
type OrderService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func NewOrderService(repos *repository.Repositories, cfg *config.Config) *OrderService {
	return &OrderService{repos: repos, cfg: cfg}
}

// CreateOrderInput represents order creation data
type CreateOrderInput struct {
	YandasID        uuid.UUID  `json:"yandas_id" binding:"required"`
	ServiceID       uuid.UUID  `json:"service_id" binding:"required"`
	AgreedPrice     float64    `json:"agreed_price" binding:"required"`
	LocationAddress string     `json:"location_address"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	ScheduledAt     *time.Time `json:"scheduled_at"`
	CustomerNotes   string     `json:"customer_notes"`
}

// Create creates a new order
func (s *OrderService) Create(customerID uuid.UUID, input *CreateOrderInput) (*models.Order, error) {
	// Verify yandaş exists and is approved
	yandas, err := s.repos.YandasProfile.GetByID(input.YandasID)
	if err != nil {
		return nil, errors.New("yandaş not found")
	}

	if yandas.ApprovalStatus != "approved" {
		return nil, errors.New("yandaş not available")
	}

	// Verify service exists
	service, err := s.repos.Service.GetByID(input.ServiceID)
	if err != nil {
		return nil, errors.New("service not found")
	}

	if service.YandasID != input.YandasID {
		return nil, errors.New("service does not belong to this yandaş")
	}

	order := &models.Order{
		CustomerID:      customerID,
		YandasID:        input.YandasID,
		ServiceID:       input.ServiceID,
		AgreedPrice:     input.AgreedPrice,
		Currency:        "TRY",
		LocationAddress: &input.LocationAddress,
		ScheduledAt:     input.ScheduledAt,
		CustomerNotes:   &input.CustomerNotes,
		Status:          "pending",
	}

	if input.Latitude != 0 {
		order.Latitude = &input.Latitude
	}
	if input.Longitude != 0 {
		order.Longitude = &input.Longitude
	}

	if err := s.repos.Order.Create(order); err != nil {
		return nil, err
	}

	return order, nil
}

// Get returns an order by ID
func (s *OrderService) Get(orderID uuid.UUID, userID uuid.UUID) (*models.Order, error) {
	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Check authorization
	if order.CustomerID != userID {
		profile, _ := s.repos.YandasProfile.GetByUserID(userID)
		if profile == nil || order.YandasID != profile.ID {
			return nil, errors.New("unauthorized")
		}
	}

	return order, nil
}

// List returns customer orders
func (s *OrderService) List(customerID uuid.UUID, page, limit int, status string) ([]models.Order, int64, error) {
	return s.repos.Order.ListByCustomer(customerID, page, limit, status)
}

// Cancel cancels an order
func (s *OrderService) Cancel(orderID uuid.UUID, userID uuid.UUID, reason string) error {
	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.CustomerID != userID {
		return errors.New("unauthorized")
	}

	if order.Status != "pending" && order.Status != "accepted" {
		return errors.New("order cannot be cancelled")
	}

	order.Status = "cancelled"
	order.CancellationReason = &reason
	order.CancelledBy = &userID

	return s.repos.Order.Update(order)
}

// ReviewInput represents review data
type ReviewInput struct {
	Rating      int    `json:"rating" binding:"required,min=1,max=5"`
	Comment     string `json:"comment"`
	IsAnonymous bool   `json:"is_anonymous"`
}

// Review adds a review to an order
func (s *OrderService) Review(orderID uuid.UUID, reviewerID uuid.UUID, input *ReviewInput) (*models.Review, error) {
	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.CustomerID != reviewerID {
		return nil, errors.New("unauthorized")
	}

	if order.Status != "completed" {
		return nil, errors.New("order must be completed to leave a review")
	}

	// Check if already reviewed
	if s.repos.Review.ExistsByOrderID(orderID) {
		return nil, errors.New("order already reviewed")
	}

	// Get yandaş user ID
	yandas, err := s.repos.YandasProfile.GetByID(order.YandasID)
	if err != nil {
		return nil, errors.New("yandaş not found")
	}

	review := &models.Review{
		OrderID:     orderID,
		ReviewerID:  reviewerID,
		RevieweeID:  yandas.UserID,
		Rating:      input.Rating,
		Comment:     &input.Comment,
		IsAnonymous: input.IsAnonymous,
	}

	if err := s.repos.Review.Create(review); err != nil {
		return nil, err
	}

	// Update yandaş rating
	s.repos.YandasProfile.UpdateRating(order.YandasID)

	return review, nil
}

// CategoryService handles category operations
type CategoryService struct {
	repos *repository.Repositories
}

func NewCategoryService(repos *repository.Repositories) *CategoryService {
	return &CategoryService{repos: repos}
}

func (s *CategoryService) List() ([]models.Category, error) {
	return s.repos.Category.List()
}

// ChatService handles chat operations
type ChatService struct {
	repos *repository.Repositories
}

func NewChatService(repos *repository.Repositories) *ChatService {
	return &ChatService{repos: repos}
}

func (s *ChatService) GetConversations(userID uuid.UUID, page, limit int) ([]models.Conversation, int64, error) {
	return s.repos.Conversation.ListByUser(userID, page, limit)
}

func (s *ChatService) GetConversation(userID uuid.UUID, convID uuid.UUID) (*models.Conversation, error) {
	conv, err := s.repos.Conversation.GetByID(convID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if conv.CustomerID != userID && conv.YandasID != userID {
		return nil, errors.New("unauthorized")
	}

	return conv, nil
}

func (s *ChatService) GetMessages(userID uuid.UUID, convID uuid.UUID, page, limit int) ([]models.Message, int64, error) {
	// Verify access
	if _, err := s.GetConversation(userID, convID); err != nil {
		return nil, 0, err
	}

	return s.repos.Message.GetByConversation(convID, page, limit)
}

// SendMessageInput represents message data
type SendMessageInput struct {
	Content     string `json:"content" binding:"required"`
	MessageType string `json:"message_type"`
}

func (s *ChatService) SendMessage(userID uuid.UUID, convID uuid.UUID, input *SendMessageInput) (*models.Message, error) {
	// Verify access
	if _, err := s.GetConversation(userID, convID); err != nil {
		return nil, err
	}

	msgType := input.MessageType
	if msgType == "" {
		msgType = "text"
	}

	msg := &models.Message{
		ConversationID: convID,
		SenderID:       userID,
		Content:        input.Content,
		MessageType:    msgType,
	}

	if err := s.repos.Message.Create(msg); err != nil {
		return nil, err
	}

	// Update conversation last message time
	s.repos.Conversation.UpdateLastMessage(convID)

	return msg, nil
}

func (s *ChatService) MarkAsRead(userID uuid.UUID, convID uuid.UUID) error {
	// Verify access
	if _, err := s.GetConversation(userID, convID); err != nil {
		return err
	}

	return s.repos.Message.MarkAsRead(convID, userID)
}

// StartConversation starts a new conversation with a yandaş
func (s *ChatService) StartConversation(customerID, yandasUserID uuid.UUID) (*models.Conversation, error) {
	return s.repos.Conversation.GetOrCreate(customerID, yandasUserID, nil)
}
