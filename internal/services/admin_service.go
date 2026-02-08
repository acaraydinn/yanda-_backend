package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
)

// AdminService handles admin operations
type AdminService struct {
	repos *repository.Repositories
}

func NewAdminService(repos *repository.Repositories) *AdminService {
	return &AdminService{repos: repos}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalUsers          int64   `json:"total_users"`
	TotalYandas         int64   `json:"total_yandas"`
	PendingApplications int64   `json:"pending_applications"`
	TotalOrders         int64   `json:"total_orders"`
	CompletedOrders     int64   `json:"completed_orders"`
	TotalRevenue        float64 `json:"total_revenue"`
	ActiveSubscriptions int64   `json:"active_subscriptions"`
}

// GetDashboard returns dashboard statistics
func (s *AdminService) GetDashboard() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// This is simplified - in production you'd have dedicated count methods
	_, total, _ := s.repos.User.List(1, 99999, "")
	stats.TotalUsers = total

	// Total yandaş
	_, yandasTotal, _ := s.repos.User.List(1, 1, "yandas")
	stats.TotalYandas = yandasTotal

	// Pending applications
	_, pendingTotal, _ := s.repos.YandasProfile.ListPendingApplications(1, 1)
	stats.PendingApplications = pendingTotal

	// Orders
	_, ordersTotal, _ := s.repos.Order.ListAll(1, 1, "")
	stats.TotalOrders = ordersTotal

	_, completedTotal, _ := s.repos.Order.ListAll(1, 1, "completed")
	stats.CompletedOrders = completedTotal

	return stats, nil
}

// ListUsers returns paginated users
func (s *AdminService) ListUsers(page, limit int, role string) ([]models.User, int64, error) {
	return s.repos.User.List(page, limit, role)
}

// GetUser returns a user by ID
func (s *AdminService) GetUser(userID uuid.UUID) (*models.User, error) {
	return s.repos.User.GetByID(userID)
}

// UpdateUser updates a user
func (s *AdminService) UpdateUser(userID uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if role, ok := updates["role"].(string); ok {
		user.Role = role
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}
	if isVerified, ok := updates["is_verified"].(bool); ok {
		user.IsVerified = isVerified
	}

	if err := s.repos.User.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *AdminService) DeleteUser(userID uuid.UUID) error {
	return s.repos.User.Delete(userID)
}

// ListApplications returns all yandaş applications
func (s *AdminService) ListApplications(page, limit int, status string) ([]models.YandasProfile, int64, error) {
	return s.repos.YandasProfile.ListAllApplications(page, limit, status)
}

// ApplicationDetailResponse wraps profile with admin-only document URLs
type ApplicationDetailResponse struct {
	*models.YandasProfile
	Documents map[string]*string `json:"documents"`
}

// GetApplication returns a yandaş application with document URLs (admin only)
func (s *AdminService) GetApplication(applicationID uuid.UUID) (*ApplicationDetailResponse, error) {
	profile, err := s.repos.YandasProfile.GetByID(applicationID)
	if err != nil {
		return nil, err
	}
	return &ApplicationDetailResponse{
		YandasProfile: profile,
		Documents: map[string]*string{
			"kimlik_on":      profile.KimlikOnURL,
			"kimlik_arka":    profile.KimlikArkaURL,
			"ehliyet_on":     profile.EhliyetOnURL,
			"ehliyet_arka":   profile.EhliyetArkaURL,
			"adli_sicil_pdf": profile.AdliSicilPDFURL,
		},
	}, nil
}

// ApproveApplication approves a yandaş application
func (s *AdminService) ApproveApplication(applicationID uuid.UUID, adminID uuid.UUID) error {
	profile, err := s.repos.YandasProfile.GetByID(applicationID)
	if err != nil {
		return err
	}

	now := time.Now()
	profile.ApprovalStatus = "approved"
	profile.ApprovedBy = &adminID
	profile.ApprovedAt = &now
	profile.IsAvailable = true

	if err := s.repos.YandasProfile.Update(profile); err != nil {
		return err
	}

	// Update user role
	user, err := s.repos.User.GetByID(profile.UserID)
	if err == nil {
		user.Role = "yandas"
		s.repos.User.Update(user)
	}

	// Log action
	s.logAction(adminID, "approve_application", "yandas_profile", applicationID, nil, map[string]interface{}{
		"status": "approved",
	})

	return nil
}

// RejectApplication rejects a yandaş application
func (s *AdminService) RejectApplication(applicationID uuid.UUID, adminID uuid.UUID, reason string) error {
	profile, err := s.repos.YandasProfile.GetByID(applicationID)
	if err != nil {
		return err
	}

	profile.ApprovalStatus = "rejected"
	profile.RejectionReason = &reason

	if err := s.repos.YandasProfile.Update(profile); err != nil {
		return err
	}

	// Log action
	s.logAction(adminID, "reject_application", "yandas_profile", applicationID, nil, map[string]interface{}{
		"status": "rejected",
		"reason": reason,
	})

	return nil
}

// ListOrders returns all orders (admin view)
func (s *AdminService) ListOrders(page, limit int, status string) ([]models.Order, int64, error) {
	return s.repos.Order.ListAll(page, limit, status)
}

// GetOrder returns an order
func (s *AdminService) GetOrder(orderID uuid.UUID) (*models.Order, error) {
	return s.repos.Order.GetByID(orderID)
}

// Category management
func (s *AdminService) CreateCategory(category *models.Category) error {
	return s.repos.Category.Create(category)
}

func (s *AdminService) UpdateCategory(category *models.Category) error {
	return s.repos.Category.Update(category)
}

func (s *AdminService) DeleteCategory(categoryID uuid.UUID) error {
	return s.repos.Category.Delete(categoryID)
}

// GetAuditLogs returns audit logs
func (s *AdminService) GetAuditLogs(page, limit int, adminID *uuid.UUID, action string) ([]models.AuditLog, int64, error) {
	return s.repos.AuditLog.List(page, limit, adminID, action)
}

func (s *AdminService) logAction(adminID uuid.UUID, action, entityType string, entityID uuid.UUID, oldValues, newValues map[string]interface{}) {
	var oldStr, newStr *string

	if oldValues != nil {
		data, _ := json.Marshal(oldValues)
		str := string(data)
		oldStr = &str
	}
	if newValues != nil {
		data, _ := json.Marshal(newValues)
		str := string(data)
		newStr = &str
	}

	log := &models.AuditLog{
		AdminID:    adminID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &entityID,
		OldValues:  oldStr,
		NewValues:  newStr,
	}

	s.repos.AuditLog.Create(log)
}

// Support Ticket methods

func (s *AdminService) ListSupportTickets(page, limit int, status, priority string) ([]models.SupportTicket, int64, error) {
	return s.repos.Support.ListTickets(page, limit, status, priority)
}

func (s *AdminService) GetSupportTicket(ticketID uuid.UUID) (*models.SupportTicket, error) {
	return s.repos.Support.GetTicket(ticketID)
}

func (s *AdminService) UpdateSupportTicket(ticketID uuid.UUID, status, priority, assignedTo string) (*models.SupportTicket, error) {
	ticket, err := s.repos.Support.GetTicket(ticketID)
	if err != nil {
		return nil, err
	}

	if status != "" {
		ticket.Status = status
		if status == "resolved" {
			now := time.Now()
			ticket.ResolvedAt = &now
		}
	}
	if priority != "" {
		ticket.Priority = priority
	}
	if assignedTo != "" {
		id, err := uuid.Parse(assignedTo)
		if err == nil {
			ticket.AssignedTo = &id
		}
	}

	err = s.repos.Support.UpdateTicket(ticket)
	return ticket, err
}

func (s *AdminService) ReplySupportTicket(ticketID, adminID uuid.UUID, content string) (*models.SupportMessage, error) {
	message := &models.SupportMessage{
		TicketID: ticketID,
		SenderID: adminID,
		Content:  content,
		IsAdmin:  true,
	}

	if err := s.repos.Support.CreateMessage(message); err != nil {
		return nil, err
	}

	// Update ticket status to pending (waiting for user response)
	ticket, _ := s.repos.Support.GetTicket(ticketID)
	if ticket != nil && ticket.Status == "open" {
		ticket.Status = "pending"
		s.repos.Support.UpdateTicket(ticket)
	}

	return message, nil
}

func (s *AdminService) GetSupportStats() (map[string]int64, error) {
	return s.repos.Support.GetStats()
}
