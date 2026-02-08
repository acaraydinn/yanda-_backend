package repository

import (
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID finds a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("YandasProfile").Preload("Subscription").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail finds a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("YandasProfile").First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone finds a user by phone
func (r *UserRepository) GetByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "phone = ?", phone).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft-deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// HardDelete permanently deletes a user (for GDPR compliance)
func (r *UserRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.User{}, "id = ?", id).Error
}

// List returns paginated users
func (r *UserRepository) List(page, limit int, role string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error

	return users, total, err
}

// ExistsByEmail checks if email exists
func (r *UserRepository) ExistsByEmail(email string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// ExistsByPhone checks if phone exists
func (r *UserRepository) ExistsByPhone(phone string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("phone = ?", phone).Count(&count)
	return count > 0
}
