package repository

import (
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// YandasProfileRepository handles yandaş profile operations
type YandasProfileRepository struct {
	db *gorm.DB
}

// NewYandasProfileRepository creates a new repository
func NewYandasProfileRepository(db *gorm.DB) *YandasProfileRepository {
	return &YandasProfileRepository{db: db}
}

// Create creates a new yandaş profile
func (r *YandasProfileRepository) Create(profile *models.YandasProfile) error {
	return r.db.Create(profile).Error
}

// GetByID finds a profile by ID
func (r *YandasProfileRepository) GetByID(id uuid.UUID) (*models.YandasProfile, error) {
	var profile models.YandasProfile
	err := r.db.Preload("User").Preload("Services").First(&profile, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetByUserID finds a profile by user ID
func (r *YandasProfileRepository) GetByUserID(userID uuid.UUID) (*models.YandasProfile, error) {
	var profile models.YandasProfile
	err := r.db.Preload("User").Preload("Services").First(&profile, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// Update updates a profile
func (r *YandasProfileRepository) Update(profile *models.YandasProfile) error {
	return r.db.Save(profile).Error
}

// ListPublic returns available and approved yandaşlar
func (r *YandasProfileRepository) ListPublic(page, limit int, categorySlug, city string) ([]models.YandasProfile, int64, error) {
	var profiles []models.YandasProfile
	var total int64

	query := r.db.Model(&models.YandasProfile{}).
		Where("approval_status = ?", "approved").
		Where("is_available = ?", true)

	if city != "" {
		query = query.Where("? = ANY(service_cities)", city)
	}

	if categorySlug != "" {
		query = query.Joins("JOIN yandas_services ON yandas_services.yandas_id = yandas_profiles.id").
			Joins("JOIN categories ON categories.id = yandas_services.category_id").
			Where("categories.slug = ?", categorySlug).
			Distinct()
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("Services.Category").
		Offset(offset).
		Limit(limit).
		Order("rating_avg DESC, total_jobs DESC").
		Find(&profiles).Error

	return profiles, total, err
}

// ListPendingApplications returns pending yandaş applications
func (r *YandasProfileRepository) ListPendingApplications(page, limit int) ([]models.YandasProfile, int64, error) {
	var profiles []models.YandasProfile
	var total int64

	query := r.db.Model(&models.YandasProfile{}).Where("approval_status = ?", "pending")
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Offset(offset).
		Limit(limit).
		Order("created_at ASC").
		Find(&profiles).Error

	return profiles, total, err
}

// ListAllApplications returns all yandaş applications (for admin)
func (r *YandasProfileRepository) ListAllApplications(page, limit int, status string) ([]models.YandasProfile, int64, error) {
	var profiles []models.YandasProfile
	var total int64

	query := r.db.Model(&models.YandasProfile{})
	if status != "" {
		query = query.Where("approval_status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&profiles).Error

	return profiles, total, err
}

// UpdateAvailability updates yandaş availability status
func (r *YandasProfileRepository) UpdateAvailability(id uuid.UUID, available bool) error {
	return r.db.Model(&models.YandasProfile{}).
		Where("id = ?", id).
		Update("is_available", available).Error
}

// UpdateLocation updates yandaş current location
func (r *YandasProfileRepository) UpdateLocation(id uuid.UUID, lat, lng float64) error {
	return r.db.Model(&models.YandasProfile{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"latitude":  lat,
			"longitude": lng,
		}).Error
}

// UpdateRating updates yandaş rating
func (r *YandasProfileRepository) UpdateRating(id uuid.UUID) error {
	// Calculate average rating from reviews
	var avgRating float64
	r.db.Model(&models.Review{}).
		Select("COALESCE(AVG(rating), 0)").
		Where("reviewee_id = (SELECT user_id FROM yandas_profiles WHERE id = ?)", id).
		Scan(&avgRating)

	var totalJobs int64
	r.db.Model(&models.Order{}).
		Where("yandas_id = ? AND status = ?", id, "completed").
		Count(&totalJobs)

	return r.db.Model(&models.YandasProfile{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"rating_avg": avgRating,
			"total_jobs": totalJobs,
		}).Error
}

// Search searches yandaş profiles by name, bio, or service title
func (r *YandasProfileRepository) Search(query string, page, limit int) ([]models.YandasProfile, int64, error) {
	var profiles []models.YandasProfile
	var total int64

	searchQuery := "%" + query + "%"
	dbQuery := r.db.Model(&models.YandasProfile{}).
		Where("approval_status = ?", "approved").
		Joins("JOIN users ON users.id = yandas_profiles.user_id").
		Where("users.full_name ILIKE ? OR yandas_profiles.bio ILIKE ?", searchQuery, searchQuery)

	dbQuery.Count(&total)

	offset := (page - 1) * limit
	err := dbQuery.
		Preload("User").
		Preload("Services.Category").
		Offset(offset).
		Limit(limit).
		Order("rating_avg DESC").
		Find(&profiles).Error

	return profiles, total, err
}
