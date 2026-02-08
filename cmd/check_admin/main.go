package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost port=5432 user=acaraydin dbname=yandas sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connect error:", err)
	}

	// Fix admin role
	result := db.Exec("UPDATE users SET role = 'admin' WHERE email = 'admin@yandas.app'")
	if result.Error != nil {
		log.Fatal("Update error:", result.Error)
	}
	fmt.Printf("✅ Admin role fixed! Rows affected: %d\n", result.RowsAffected)

	// Verify
	var role string
	db.Raw("SELECT role FROM users WHERE email = 'admin@yandas.app'").Scan(&role)
	fmt.Printf("✅ Admin role is now: %s\n", role)
}
