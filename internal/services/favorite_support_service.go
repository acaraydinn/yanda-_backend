package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
)

// FavoriteService handles favorite operations
type FavoriteService struct {
	repos *repository.Repositories
}

// NewFavoriteService creates a new favorite service
func NewFavoriteService(repos *repository.Repositories) *FavoriteService {
	return &FavoriteService{repos: repos}
}

// Toggle adds or removes a yandaş from favorites
func (s *FavoriteService) Toggle(userID, yandasID uuid.UUID) (bool, error) {
	// Verify yandaş exists
	_, err := s.repos.YandasProfile.GetByID(yandasID)
	if err != nil {
		return false, errors.New("yandaş not found")
	}

	// Check if already favorited
	if s.repos.Favorite.Exists(userID, yandasID) {
		// Remove from favorites
		if err := s.repos.Favorite.Delete(userID, yandasID); err != nil {
			return false, err
		}
		return false, nil
	}

	// Add to favorites
	fav := &models.Favorite{
		UserID:   userID,
		YandasID: yandasID,
	}

	if err := s.repos.Favorite.Create(fav); err != nil {
		return false, err
	}

	return true, nil
}

// List returns user favorites
func (s *FavoriteService) List(userID uuid.UUID, page, limit int) ([]models.Favorite, int64, error) {
	return s.repos.Favorite.ListByUser(userID, page, limit)
}

// IsFavorited checks if a yandaş is favorited by user
func (s *FavoriteService) IsFavorited(userID, yandasID uuid.UUID) bool {
	return s.repos.Favorite.Exists(userID, yandasID)
}

// GetFavoriteIDs returns all favorited yandaş IDs for a user
func (s *FavoriteService) GetFavoriteIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	return s.repos.Favorite.GetYandasIDs(userID)
}

// SupportService handles user-facing support operations
type SupportService struct {
	repos *repository.Repositories
}

// NewSupportService creates a new support service
func NewSupportService(repos *repository.Repositories) *SupportService {
	return &SupportService{repos: repos}
}

// CreateTicketInput represents support ticket creation data
type CreateTicketInput struct {
	Subject     string `json:"subject" binding:"required"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category"` // general, order, payment, account, technical
	OrderID     string `json:"order_id"`
}

// CreateTicket creates a new support ticket
func (s *SupportService) CreateTicket(userID uuid.UUID, input *CreateTicketInput) (*models.SupportTicket, error) {
	category := input.Category
	if category == "" {
		category = "general"
	}

	ticket := &models.SupportTicket{
		UserID:      userID,
		Subject:     input.Subject,
		Description: input.Description,
		Category:    category,
		Priority:    "normal",
		Status:      "open",
	}

	if input.OrderID != "" {
		orderID, err := uuid.Parse(input.OrderID)
		if err == nil {
			ticket.OrderID = &orderID
		}
	}

	if err := s.repos.Support.CreateTicket(ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

// ListUserTickets returns support tickets for a user
func (s *SupportService) ListUserTickets(userID uuid.UUID, page, limit int) ([]models.SupportTicket, int64, error) {
	return s.repos.Support.ListByUser(userID, page, limit)
}

// GetUserTicket returns a specific support ticket (user-scoped)
func (s *SupportService) GetUserTicket(userID uuid.UUID, ticketID uuid.UUID) (*models.SupportTicket, error) {
	ticket, err := s.repos.Support.GetTicket(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	if ticket.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return ticket, nil
}

// ReplyTicket adds a reply to a support ticket
func (s *SupportService) ReplyTicket(userID uuid.UUID, ticketID uuid.UUID, content string) (*models.SupportMessage, error) {
	ticket, err := s.repos.Support.GetTicket(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	if ticket.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	msg := &models.SupportMessage{
		TicketID: ticketID,
		SenderID: userID,
		Content:  content,
		IsAdmin:  false,
	}

	if err := s.repos.Support.CreateMessage(msg); err != nil {
		return nil, err
	}

	return msg, nil
}
