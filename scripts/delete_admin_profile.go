package main

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/yandas?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	// Delete admin's yandas profile
	adminUserID := "6ca6d14f-36d1-4d22-9c03-79dbe80e805d"
	result := db.Exec("DELETE FROM yandas_profiles WHERE user_id = ?", adminUserID)
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
		return
	}
	fmt.Printf("Deleted %d yandas profile(s) for admin user\n", result.RowsAffected)
}
