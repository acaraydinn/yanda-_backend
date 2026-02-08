package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yandas/backend/internal/models"
)

func main() {
	godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB bağlantısı başarısız:", err)
	}

	// Admin kullanıcısını bul ve rolünü güncelle
	var user models.User
	if err := db.Where("email = ?", "admin@yandas.app").First(&user).Error; err != nil {
		log.Fatal("Kullanıcı bulunamadı:", err)
	}

	user.Role = "yandas"
	db.Save(&user)
	fmt.Printf("✅ %s (%s) kullanıcısının rolü 'yandas' olarak güncellendi!\n", user.FullName, user.Email)

	// Yandaş profili kontrol et, hizmet ekle
	var profile models.YandasProfile
	if err := db.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		log.Fatal("Yandaş profili bulunamadı:", err)
	}

	// Hizmetleri kontrol et
	var serviceCount int64
	db.Model(&models.YandasService{}).Where("yandas_id = ?", profile.ID).Count(&serviceCount)
	if serviceCount == 0 {
		var categories []models.Category
		db.Limit(3).Find(&categories)
		for _, cat := range categories {
			desc := cat.Name + " alanında profesyonel hizmet sunuyorum."
			dur := 60
			service := models.YandasService{
				YandasID:        profile.ID,
				CategoryID:      cat.ID,
				Title:           cat.Name + " Hizmeti",
				Description:     &desc,
				BasePrice:       150.0,
				DurationMinutes: &dur,
				IsActive:        true,
			}
			db.Create(&service)
			fmt.Printf("  ✅ Hizmet eklendi: %s (%.0f TL)\n", service.Title, service.BasePrice)
		}
	} else {
		fmt.Printf("  ℹ️  Zaten %d hizmet mevcut\n", serviceCount)
	}

	fmt.Println("\n🎉 admin@yandas.app artık aktif bir Yandaş!")
}
