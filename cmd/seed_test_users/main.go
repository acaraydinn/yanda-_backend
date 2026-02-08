package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	dsn := getEnv("DATABASE_URL", "postgres://acaraydin@localhost:5432/yandas?sslmode=disable")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	log.Println("🗑️  Deleting all user-related data...")

	// Delete in correct order (foreign keys)
	db.Exec("DELETE FROM call_logs")
	db.Exec("DELETE FROM messages")
	db.Exec("DELETE FROM conversations")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM reviews")
	db.Exec("DELETE FROM yandas_services")
	db.Exec("DELETE FROM subscriptions")
	db.Exec("DELETE FROM device_tokens")
	db.Exec("DELETE FROM notifications")
	db.Exec("DELETE FROM yandas_profiles")
	db.Exec("DELETE FROM support_tickets")
	db.Exec("DELETE FROM users WHERE role != 'admin'")

	log.Println("✅ All user data deleted!")

	// Hash password
	password := "Test1234!"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hash)

	// Create User 1 - Hakan
	user1ID := uuid.New()
	email1 := "hakan@test.com"
	phone1 := "5551112233"
	db.Exec(`INSERT INTO users (id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user1ID, email1, phone1, hashStr, "Hakan Aydın", "yandas", true, true, time.Now(), time.Now())

	// Create User 2 - Fevzi
	user2ID := uuid.New()
	email2 := "fevzi@test.com"
	phone2 := "5554445566"
	db.Exec(`INSERT INTO users (id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user2ID, email2, phone2, hashStr, "Fevzi Acar", "yandas", true, true, time.Now(), time.Now())

	log.Printf("✅ Created user 1: Hakan Aydın (ID: %s)", user1ID)
	log.Printf("✅ Created user 2: Fevzi Acar (ID: %s)", user2ID)

	// Create YandasProfile for both
	now := time.Now()
	profile1ID := uuid.New()
	bio1 := "Profesyonel yandaş hizmetleri. Güvenilir ve hızlı."
	cities1 := "{İstanbul,Ankara}"
	db.Exec(`INSERT INTO yandas_profiles (id, user_id, bio, approval_status, approved_at, rating_avg, total_jobs, is_available, service_cities, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?::text[], ?)`,
		profile1ID, user1ID, bio1, "approved", now, 4.8, 50, true, cities1, now)

	profile2ID := uuid.New()
	bio2 := "Deneyimli yandaş. Her işte yanındayım."
	cities2 := "{İstanbul,İzmir}"
	db.Exec(`INSERT INTO yandas_profiles (id, user_id, bio, approval_status, approved_at, rating_avg, total_jobs, is_available, service_cities, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?::text[], ?)`,
		profile2ID, user2ID, bio2, "approved", now, 4.9, 75, true, cities2, now)

	log.Println("✅ YandasProfiles created with 'approved' status!")

	// Get all sub-categories (leaf categories with a parent_id)
	type Cat struct {
		ID   uuid.UUID
		Name string
	}
	var subCats []Cat
	db.Raw("SELECT id, name FROM categories WHERE parent_id IS NOT NULL ORDER BY sort_order").Scan(&subCats)

	log.Printf("📋 Found %d sub-categories to assign as services", len(subCats))

	// Create YandasService for each sub-category for both profiles
	for _, cat := range subCats {
		// For Hakan
		db.Exec(`INSERT INTO yandas_services (id, yandas_id, category_id, title, description, base_price, currency, duration_minutes, is_active, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid.New(), profile1ID, cat.ID, cat.Name+" Hizmeti", cat.Name+" için profesyonel hizmet", 500.00, "TRY", 60, true, now)

		// For Fevzi
		db.Exec(`INSERT INTO yandas_services (id, yandas_id, category_id, title, description, base_price, currency, duration_minutes, is_active, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid.New(), profile2ID, cat.ID, cat.Name+" Hizmeti", cat.Name+" için profesyonel hizmet", 450.00, "TRY", 60, true, now)
	}

	log.Printf("✅ Created %d services for each yandaş (%d total)", len(subCats), len(subCats)*2)

	fmt.Println("")
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              TEST HESAPLARI OLUŞTURULDU                 ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║  👤 Hakan Aydın                                        ║\n")
	fmt.Printf("║     📧 Email: %-40s ║\n", email1)
	fmt.Printf("║     🔑 Şifre: %-40s ║\n", password)
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║  👤 Fevzi Acar                                         ║\n")
	fmt.Printf("║     📧 Email: %-40s ║\n", email2)
	fmt.Printf("║     🔑 Şifre: %-40s ║\n", password)
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║  ✅ Yandaş onayı: Her iki hesap da 'approved'          ║\n")
	fmt.Printf("║  📋 Hizmetler: %d alt kategori x 2 hesap = %d hizmet    ║\n", len(subCats), len(subCats)*2)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
