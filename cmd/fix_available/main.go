package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type YandasProfile struct {
	ID             string `gorm:"column:id"`
	UserID         string `gorm:"column:user_id"`
	ApprovalStatus string `gorm:"column:approval_status"`
	IsAvailable    bool   `gorm:"column:is_available"`
}

func (YandasProfile) TableName() string { return "yandas_profiles" }

func main() {
	dsn := "host=localhost user=acaraydin dbname=yandas port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	var profiles []YandasProfile
	db.Find(&profiles)
	for _, p := range profiles {
		fmt.Printf("ID: %s | Status: %s | Available: %v\n", p.ID[:8], p.ApprovalStatus, p.IsAvailable)
	}

	// Fix all approved profiles to be available
	result := db.Model(&YandasProfile{}).Where("approval_status = ? AND is_available = ?", "approved", false).Update("is_available", true)
	fmt.Printf("\nFixed %d approved profiles to is_available=true\n", result.RowsAffected)
}
