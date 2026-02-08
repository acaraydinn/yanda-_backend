package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// ConversationRepository handles conversation operations
type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(conv *models.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *ConversationRepository) GetByID(id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.
		Preload("Customer").
		Preload("Yandas").
		First(&conv, "id = ?", id).Error
	return &conv, err
}

func (r *ConversationRepository) GetByParticipants(customerID, yandasID uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.
		Where("customer_id = ? AND yandas_id = ?", customerID, yandasID).
		First(&conv).Error
	return &conv, err
}

func (r *ConversationRepository) GetOrCreate(customerID, yandasID uuid.UUID, orderID *uuid.UUID) (*models.Conversation, error) {
	conv, err := r.GetByParticipants(customerID, yandasID)
	if err == nil {
		// Re-fetch with preloads
		return r.GetByID(conv.ID)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	conv = &models.Conversation{
		CustomerID: customerID,
		YandasID:   yandasID,
		OrderID:    orderID,
	}

	if err := r.Create(conv); err != nil {
		return nil, err
	}

	// Re-fetch with preloads
	return r.GetByID(conv.ID)
}

func (r *ConversationRepository) ListByUser(userID uuid.UUID, page, limit int) ([]models.Conversation, int64, error) {
	var convs []models.Conversation
	var total int64

	query := r.db.Model(&models.Conversation{}).
		Where("customer_id = ? OR yandas_id = ?", userID, userID)

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Customer").
		Preload("Yandas").
		Offset(offset).
		Limit(limit).
		Order("last_message_at DESC NULLS LAST").
		Find(&convs).Error

	return convs, total, err
}

func (r *ConversationRepository) UpdateLastMessage(id uuid.UUID) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("last_message_at", time.Now()).Error
}

// MessageRepository handles message operations
type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *models.Message) error {
	return r.db.Create(msg).Error
}

func (r *MessageRepository) GetByConversation(conversationID uuid.UUID, page, limit int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := r.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Sender").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&messages).Error

	return messages, total, err
}

func (r *MessageRepository) MarkAsRead(conversationID, userID uuid.UUID) error {
	return r.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ?", conversationID, userID, false).
		Update("is_read", true).Error
}

func (r *MessageRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Where("(conversations.customer_id = ? OR conversations.yandas_id = ?)", userID, userID).
		Where("messages.sender_id != ? AND messages.is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
