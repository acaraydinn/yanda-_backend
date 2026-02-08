package repository

import (
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// FavoriteRepository handles favorite operations
type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Create(fav *models.Favorite) error {
	return r.db.Create(fav).Error
}

func (r *FavoriteRepository) Delete(userID, yandasID uuid.UUID) error {
	return r.db.Where("user_id = ? AND yandas_id = ?", userID, yandasID).Delete(&models.Favorite{}).Error
}

func (r *FavoriteRepository) Exists(userID, yandasID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.Favorite{}).Where("user_id = ? AND yandas_id = ?", userID, yandasID).Count(&count)
	return count > 0
}

func (r *FavoriteRepository) ListByUser(userID uuid.UUID, page, limit int) ([]models.Favorite, int64, error) {
	var favs []models.Favorite
	var total int64

	query := r.db.Model(&models.Favorite{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Yandas").
		Preload("Yandas.User").
		Preload("Yandas.Services").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&favs).Error

	return favs, total, err
}

func (r *FavoriteRepository) GetYandasIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&models.Favorite{}).Where("user_id = ?", userID).Pluck("yandas_id", &ids).Error
	return ids, err
}
