package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"gorm.io/gorm"
)

// OrderRepository handles order operations
type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	// Generate order number
	order.OrderNumber = generateOrderNumber()
	return r.db.Create(order).Error
}

func (r *OrderRepository) GetByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.
		Preload("Customer").
		Preload("Yandas.User").
		Preload("Service.Category").
		Preload("Review").
		First(&order, "id = ?", id).Error
	return &order, err
}

func (r *OrderRepository) GetByOrderNumber(orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.
		Preload("Customer").
		Preload("Yandas.User").
		Preload("Service").
		First(&order, "order_number = ?", orderNumber).Error
	return &order, err
}

func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) ListByCustomer(customerID uuid.UUID, page, limit int, status string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("customer_id = ?", customerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Yandas.User").
		Preload("Service").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) ListByYandas(yandasID uuid.UUID, page, limit int, status string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("yandas_id = ?", yandasID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Customer").
		Preload("Service").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) ListAll(page, limit int, status string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Customer").
		Preload("Yandas.User").
		Preload("Service").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) UpdateStatus(id uuid.UUID, status string) error {
	updates := map[string]interface{}{"status": status}

	switch status {
	case "in_progress":
		updates["started_at"] = time.Now()
	case "completed":
		updates["completed_at"] = time.Now()
	}

	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OrderRepository) GetStats(yandasID uuid.UUID) (map[string]interface{}, error) {
	var stats struct {
		TotalOrders     int64   `json:"total_orders"`
		CompletedOrders int64   `json:"completed_orders"`
		TotalRevenue    float64 `json:"total_revenue"`
		AvgRating       float64 `json:"avg_rating"`
	}

	r.db.Model(&models.Order{}).
		Where("yandas_id = ?", yandasID).
		Count(&stats.TotalOrders)

	r.db.Model(&models.Order{}).
		Where("yandas_id = ? AND status = ?", yandasID, "completed").
		Count(&stats.CompletedOrders)

	r.db.Model(&models.Order{}).
		Where("yandas_id = ? AND status = ?", yandasID, "completed").
		Select("COALESCE(SUM(agreed_price), 0)").
		Scan(&stats.TotalRevenue)

	return map[string]interface{}{
		"total_orders":     stats.TotalOrders,
		"completed_orders": stats.CompletedOrders,
		"total_revenue":    stats.TotalRevenue,
	}, nil
}

func generateOrderNumber() string {
	return fmt.Sprintf("YND%d%04d", time.Now().Unix()%100000, time.Now().Nanosecond()%10000)
}

// ReviewRepository handles review operations
type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *ReviewRepository) GetByOrderID(orderID uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := r.db.Preload("Reviewer").First(&review, "order_id = ?", orderID).Error
	return &review, err
}

func (r *ReviewRepository) ListByReviewee(revieweeID uuid.UUID, page, limit int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	query := r.db.Model(&models.Review{}).Where("reviewee_id = ?", revieweeID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Reviewer").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&reviews).Error

	return reviews, total, err
}

func (r *ReviewRepository) ExistsByOrderID(orderID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.Review{}).Where("order_id = ?", orderID).Count(&count)
	return count > 0
}
