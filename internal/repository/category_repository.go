package repository

import (
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// CategoryRepository handles category operations
type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List() ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Where("is_active = ? AND parent_id IS NULL", true).
		Preload("SubCategories", "is_active = ?", true).
		Order("sort_order ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) GetByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, "id = ?", id).Error
	return &category, err
}

func (r *CategoryRepository) GetBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, "slug = ?", slug).Error
	return &category, err
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&models.Category{}).Where("id = ?", id).Update("is_active", false).Error
}

// ServiceRepository handles yandaş service operations
type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(service *models.YandasService) error {
	return r.db.Create(service).Error
}

func (r *ServiceRepository) GetByID(id uuid.UUID) (*models.YandasService, error) {
	var service models.YandasService
	err := r.db.Preload("Category").First(&service, "id = ?", id).Error
	return &service, err
}

func (r *ServiceRepository) GetByYandasID(yandasID uuid.UUID) ([]models.YandasService, error) {
	var services []models.YandasService
	err := r.db.Preload("Category").Where("yandas_id = ? AND is_active = ?", yandasID, true).Find(&services).Error
	return services, err
}

func (r *ServiceRepository) Update(service *models.YandasService) error {
	return r.db.Save(service).Error
}

func (r *ServiceRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&models.YandasService{}).Where("id = ?", id).Update("is_active", false).Error
}
