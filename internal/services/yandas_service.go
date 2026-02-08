package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/repository"
)

// YandasService handles yandaş operations
type YandasService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

// NewYandasService creates a new yandaş service
func NewYandasService(repos *repository.Repositories, cfg *config.Config) *YandasService {
	return &YandasService{repos: repos, cfg: cfg}
}

// ApplicationInput represents yandaş application data
type ApplicationInput struct {
	Bio             string   `json:"bio"`
	InstagramHandle string   `json:"instagram_handle" binding:"required"`
	KimlikOnURL     string   `json:"kimlik_on_url"`      // ID Card Front
	KimlikArkaURL   string   `json:"kimlik_arka_url"`    // ID Card Back
	EhliyetOnURL    string   `json:"ehliyet_on_url"`     // License Front
	EhliyetArkaURL  string   `json:"ehliyet_arka_url"`   // License Back
	AdliSicilPDFURL string   `json:"adli_sicil_pdf_url"` // Criminal Record PDF
	ServiceCities   []string `json:"service_cities" binding:"required"`
	CategoryIDs     []string `json:"category_ids"` // For multiple categories (optional for now)
}

// Apply creates a yandaş application
func (s *YandasService) Apply(userID uuid.UUID, input *ApplicationInput) (*models.YandasProfile, error) {
	// Check if already applied
	existing, _ := s.repos.YandasProfile.GetByUserID(userID)
	if existing != nil {
		return nil, errors.New("you have already applied to become a yandaş")
	}

	profile := &models.YandasProfile{
		UserID:          userID,
		Bio:             &input.Bio,
		InstagramHandle: &input.InstagramHandle,
		ServiceCities:   input.ServiceCities,
		ApprovalStatus:  "pending",
	}

	// Set document URLs if provided (Front and Back for ID and License)
	if input.KimlikOnURL != "" {
		profile.KimlikOnURL = &input.KimlikOnURL
	}
	if input.KimlikArkaURL != "" {
		profile.KimlikArkaURL = &input.KimlikArkaURL
	}
	if input.EhliyetOnURL != "" {
		profile.EhliyetOnURL = &input.EhliyetOnURL
	}
	if input.EhliyetArkaURL != "" {
		profile.EhliyetArkaURL = &input.EhliyetArkaURL
	}
	if input.AdliSicilPDFURL != "" {
		profile.AdliSicilPDFURL = &input.AdliSicilPDFURL
	}

	if err := s.repos.YandasProfile.Create(profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// GetApplicationStatus returns application status
func (s *YandasService) GetApplicationStatus(userID uuid.UUID) (*models.YandasProfile, error) {
	return s.repos.YandasProfile.GetByUserID(userID)
}

// UpdateProfileInput represents yandaş profile update data
type UpdateYandasProfileInput struct {
	Bio           string   `json:"bio"`
	ServiceCities []string `json:"service_cities"`
}

// UpdateProfile updates yandaş profile
func (s *YandasService) UpdateProfile(userID uuid.UUID, input *UpdateYandasProfileInput) (*models.YandasProfile, error) {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("yandaş profile not found")
	}

	if profile.ApprovalStatus != "approved" {
		return nil, errors.New("profile not approved yet")
	}

	if input.Bio != "" {
		profile.Bio = &input.Bio
	}

	if len(input.ServiceCities) > 0 {
		profile.ServiceCities = input.ServiceCities
	}

	if err := s.repos.YandasProfile.Update(profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// UpdateAvailability updates availability status
func (s *YandasService) UpdateAvailability(userID uuid.UUID, available bool) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	if profile.ApprovalStatus != "approved" {
		return errors.New("profile not approved yet")
	}

	return s.repos.YandasProfile.UpdateAvailability(profile.ID, available)
}

// UpdateLocation updates current location
func (s *YandasService) UpdateLocation(userID uuid.UUID, lat, lng float64) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	return s.repos.YandasProfile.UpdateLocation(profile.ID, lat, lng)
}

// ListPublic returns available yandaşlar
func (s *YandasService) ListPublic(page, limit int, category, city string) ([]models.YandasProfile, int64, error) {
	return s.repos.YandasProfile.ListPublic(page, limit, category, city)
}

// GetPublic returns a public yandaş profile
func (s *YandasService) GetPublic(id uuid.UUID) (*models.YandasProfile, error) {
	profile, err := s.repos.YandasProfile.GetByID(id)
	if err != nil {
		return nil, err
	}

	if profile.ApprovalStatus != "approved" {
		return nil, errors.New("profile not found")
	}

	return profile, nil
}

// GetServices returns yandaş services
func (s *YandasService) GetServices(yandasID uuid.UUID) ([]models.YandasService, error) {
	return s.repos.Service.GetByYandasID(yandasID)
}

// GetReviews returns yandaş reviews
func (s *YandasService) GetReviews(yandasID uuid.UUID, page, limit int) ([]models.Review, int64, error) {
	profile, err := s.repos.YandasProfile.GetByID(yandasID)
	if err != nil {
		return nil, 0, err
	}

	return s.repos.Review.ListByReviewee(profile.UserID, page, limit)
}

// ServiceInput represents service creation data
type ServiceInput struct {
	CategoryID      uuid.UUID `json:"category_id" binding:"required"`
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	BasePrice       float64   `json:"base_price" binding:"required"`
	DurationMinutes int       `json:"duration_minutes"`
	Includes        []string  `json:"includes"`
}

// CreateService creates a new service
func (s *YandasService) CreateService(userID uuid.UUID, input *ServiceInput) (*models.YandasService, error) {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("yandaş profile not found")
	}

	if profile.ApprovalStatus != "approved" {
		return nil, errors.New("profile not approved yet")
	}

	service := &models.YandasService{
		YandasID:        profile.ID,
		CategoryID:      input.CategoryID,
		Title:           input.Title,
		Description:     &input.Description,
		BasePrice:       input.BasePrice,
		DurationMinutes: &input.DurationMinutes,
		Includes:        input.Includes,
		IsActive:        true,
	}

	if err := s.repos.Service.Create(service); err != nil {
		return nil, err
	}

	return service, nil
}

// UpdateService updates a service
func (s *YandasService) UpdateService(userID uuid.UUID, serviceID uuid.UUID, input *ServiceInput) (*models.YandasService, error) {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("yandaş profile not found")
	}

	service, err := s.repos.Service.GetByID(serviceID)
	if err != nil {
		return nil, errors.New("service not found")
	}

	if service.YandasID != profile.ID {
		return nil, errors.New("unauthorized")
	}

	service.Title = input.Title
	service.Description = &input.Description
	service.BasePrice = input.BasePrice
	service.DurationMinutes = &input.DurationMinutes
	service.Includes = input.Includes

	if err := s.repos.Service.Update(service); err != nil {
		return nil, err
	}

	return service, nil
}

// DeleteService deletes a service
func (s *YandasService) DeleteService(userID uuid.UUID, serviceID uuid.UUID) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	service, err := s.repos.Service.GetByID(serviceID)
	if err != nil {
		return errors.New("service not found")
	}

	if service.YandasID != profile.ID {
		return errors.New("unauthorized")
	}

	return s.repos.Service.Delete(serviceID)
}

// GetOrders returns yandaş orders
func (s *YandasService) GetOrders(userID uuid.UUID, page, limit int, status string) ([]models.Order, int64, error) {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return nil, 0, errors.New("yandaş profile not found")
	}

	return s.repos.Order.ListByYandas(profile.ID, page, limit, status)
}

// AcceptOrder accepts an order
func (s *YandasService) AcceptOrder(userID uuid.UUID, orderID uuid.UUID) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.YandasID != profile.ID {
		return errors.New("unauthorized")
	}

	if order.Status != "pending" {
		return errors.New("order cannot be accepted")
	}

	return s.repos.Order.UpdateStatus(orderID, "accepted")
}

// RejectOrder rejects an order
func (s *YandasService) RejectOrder(userID uuid.UUID, orderID uuid.UUID, reason string) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.YandasID != profile.ID {
		return errors.New("unauthorized")
	}

	if order.Status != "pending" {
		return errors.New("order cannot be rejected")
	}

	order.Status = "cancelled"
	order.CancellationReason = &reason
	order.CancelledBy = &profile.UserID
	return s.repos.Order.Update(order)
}

// StartOrder starts an order
func (s *YandasService) StartOrder(userID uuid.UUID, orderID uuid.UUID) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.YandasID != profile.ID {
		return errors.New("unauthorized")
	}

	if order.Status != "accepted" {
		return errors.New("order cannot be started")
	}

	return s.repos.Order.UpdateStatus(orderID, "in_progress")
}

// CompleteOrder completes an order
func (s *YandasService) CompleteOrder(userID uuid.UUID, orderID uuid.UUID, notes string) error {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return errors.New("yandaş profile not found")
	}

	order, err := s.repos.Order.GetByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.YandasID != profile.ID {
		return errors.New("unauthorized")
	}

	if order.Status != "in_progress" {
		return errors.New("order cannot be completed")
	}

	now := time.Now()
	order.Status = "completed"
	order.CompletedAt = &now
	order.YandasNotes = &notes

	if err := s.repos.Order.Update(order); err != nil {
		return err
	}

	// Update yandaş rating
	s.repos.YandasProfile.UpdateRating(profile.ID)

	return nil
}

// GetStats returns yandaş stats
func (s *YandasService) GetStats(userID uuid.UUID) (map[string]interface{}, error) {
	profile, err := s.repos.YandasProfile.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("yandaş profile not found")
	}

	stats, err := s.repos.Order.GetStats(profile.ID)
	if err != nil {
		return nil, err
	}

	stats["rating_avg"] = profile.RatingAvg
	stats["total_jobs"] = profile.TotalJobs

	return stats, nil
}

// Search searches yandaş profiles by query
func (s *YandasService) Search(query string, page, limit int) ([]models.YandasProfile, int64, error) {
	return s.repos.YandasProfile.Search(query, page, limit)
}
