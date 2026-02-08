package repository

import (
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

type SupportRepository struct {
	db *gorm.DB
}

func NewSupportRepository(db *gorm.DB) *SupportRepository {
	return &SupportRepository{db: db}
}

func (r *SupportRepository) ListTickets(page, limit int, status, priority string) ([]models.SupportTicket, int64, error) {
	var tickets []models.SupportTicket
	var total int64

	query := r.db.Model(&models.SupportTicket{}).Preload("User").Preload("Assignee")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tickets).Error
	return tickets, total, err
}

func (r *SupportRepository) GetTicket(id uuid.UUID) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	err := r.db.Preload("User").Preload("Assignee").Preload("Messages.Sender").First(&ticket, "id = ?", id).Error
	return &ticket, err
}

func (r *SupportRepository) CreateTicket(ticket *models.SupportTicket) error {
	return r.db.Create(ticket).Error
}

func (r *SupportRepository) UpdateTicket(ticket *models.SupportTicket) error {
	return r.db.Save(ticket).Error
}

func (r *SupportRepository) CreateMessage(message *models.SupportMessage) error {
	return r.db.Create(message).Error
}

func (r *SupportRepository) GetTicketMessages(ticketID uuid.UUID) ([]models.SupportMessage, error) {
	var messages []models.SupportMessage
	err := r.db.Preload("Sender").Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

func (r *SupportRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)
	var open, pending, resolved, urgent, total int64

	r.db.Model(&models.SupportTicket{}).Where("status = ?", "open").Count(&open)
	r.db.Model(&models.SupportTicket{}).Where("status = ?", "pending").Count(&pending)
	r.db.Model(&models.SupportTicket{}).Where("status = ?", "resolved").Count(&resolved)
	r.db.Model(&models.SupportTicket{}).Where("priority = ?", "urgent").Where("status != ?", "resolved").Count(&urgent)
	r.db.Model(&models.SupportTicket{}).Count(&total)

	stats["open"] = open
	stats["pending"] = pending
	stats["resolved"] = resolved
	stats["urgent"] = urgent
	stats["total"] = total

	return stats, nil
}

func (r *SupportRepository) ListByUser(userID uuid.UUID, page, limit int) ([]models.SupportTicket, int64, error) {
	var tickets []models.SupportTicket
	var total int64

	query := r.db.Model(&models.SupportTicket{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tickets).Error
	return tickets, total, err
}
