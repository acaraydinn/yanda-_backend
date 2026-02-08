package database

import (
	"log"

	"github.com/google/uuid"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed populates the database with initial data
func Seed(db *gorm.DB, cfg *config.Config) error {
	log.Println("🌱 Seeding database...")

	// Seed admin user
	if err := seedAdmin(db, cfg); err != nil {
		return err
	}

	// Seed categories
	if err := seedCategories(db); err != nil {
		return err
	}

	log.Println("✅ Database seeding completed")
	return nil
}

func seedAdmin(db *gorm.DB, cfg *config.Config) error {
	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return nil // Admin already exists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		ID:           uuid.New(),
		Email:        &cfg.AdminEmail,
		PasswordHash: string(hashedPassword),
		FullName:     "Admin",
		Role:         "admin",
		IsVerified:   true,
		IsActive:     true,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Printf("Failed to create admin: %v", err)
		return nil // Non-fatal, might already exist
	}

	log.Printf("✅ Admin user created: %s", cfg.AdminEmail)
	return nil
}

func seedCategories(db *gorm.DB) error {
	// Check if we already have subcategories (parent_id not null)
	var subCount int64
	db.Model(&models.Category{}).Where("parent_id IS NOT NULL").Count(&subCount)
	if subCount > 0 {
		return nil // Already seeded with subcategories
	}

	// Nullify category references in yandas_services to avoid FK constraint
	db.Exec("UPDATE yandas_services SET category_id = NULL WHERE category_id IS NOT NULL")
	// Delete ALL existing categories (raw SQL to bypass soft-delete)
	db.Exec("DELETE FROM categories")

	type subCat struct {
		Name   string
		NameEN string
		Slug   string
	}

	type mainCat struct {
		Name   string
		NameEN string
		Slug   string
		Icon   string
		Desc   string
		Subs   []subCat
	}

	categories := []mainCat{
		{
			Name: "Vasıta & Ekspertiz", NameEN: "Vehicle & Expertise",
			Slug: "vasita-ekspertiz", Icon: "car",
			Desc: "Araç alımında yerinde kontrol ve ekspertiz refakati",
			Subs: []subCat{
				{Name: "Araç Ekspertiz Refakati", NameEN: "Vehicle Expertise Escort", Slug: "arac-ekspertiz-refakati"},
				{Name: "Araç Alımda Kontrol", NameEN: "Vehicle Purchase Inspection", Slug: "arac-alimda-kontrol"},
				{Name: "Araç Teslimat", NameEN: "Vehicle Delivery", Slug: "arac-teslimat"},
			},
		},
		{
			Name: "Vekil Sürücü", NameEN: "Proxy Driver",
			Slug: "vekil-surucu", Icon: "steering-wheel",
			Desc: "Güvenli valelik ve şoför hizmeti",
			Subs: []subCat{
				{Name: "Şoför Hizmeti", NameEN: "Chauffeur Service", Slug: "sofor-hizmeti"},
				{Name: "Valet Hizmeti", NameEN: "Valet Service", Slug: "valet-hizmeti"},
				{Name: "Uzun Yol Sürücü", NameEN: "Long Distance Driver", Slug: "uzun-yol-surucu"},
			},
		},
		{
			Name: "Kurye & Lojistik", NameEN: "Courier & Logistics",
			Slug: "kurye-lojistik", Icon: "package",
			Desc: "Belge ve paket teslim hizmetleri",
			Subs: []subCat{
				{Name: "Belge Teslimi", NameEN: "Document Delivery", Slug: "belge-teslimi"},
				{Name: "Paket Taşıma", NameEN: "Package Transport", Slug: "paket-tasima"},
				{Name: "Market Alışverişi", NameEN: "Grocery Shopping", Slug: "market-alisverisi"},
			},
		},
		{
			Name: "Eşya Kontrolü", NameEN: "Item Inspection",
			Slug: "esya-kontrolu", Icon: "search",
			Desc: "İkinci el eşya yerinde kontrol",
			Subs: []subCat{
				{Name: "İkinci El Eşya Kontrol", NameEN: "Second-hand Item Inspection", Slug: "ikinci-el-esya-kontrol"},
				{Name: "Emlak Kontrol", NameEN: "Real Estate Inspection", Slug: "emlak-kontrol"},
				{Name: "Teknoloji Ürün Kontrol", NameEN: "Tech Product Inspection", Slug: "teknoloji-urun-kontrol"},
			},
		},
		{
			Name: "Kişisel Asistanlık", NameEN: "Personal Assistance",
			Slug: "kisisel-asistanlik", Icon: "user",
			Desc: "Günlük işlerde vekil yardım",
			Subs: []subCat{
				{Name: "Randevu Takibi", NameEN: "Appointment Follow-up", Slug: "randevu-takibi"},
				{Name: "Kuyruk Bekleme", NameEN: "Queue Waiting", Slug: "kuyruk-bekleme"},
				{Name: "Resmi İşlem Vekili", NameEN: "Official Procedure Proxy", Slug: "resmi-islem-vekili"},
			},
		},
	}

	for i, mc := range categories {
		parentID := uuid.New()
		parent := models.Category{
			ID:          parentID,
			Name:        mc.Name,
			NameEN:      strPtr(mc.NameEN),
			Slug:        mc.Slug,
			Icon:        strPtr(mc.Icon),
			Description: strPtr(mc.Desc),
			SortOrder:   i + 1,
		}

		if err := db.Create(&parent).Error; err != nil {
			log.Printf("Failed to create category %s: %v", mc.Name, err)
			continue
		}

		for j, sc := range mc.Subs {
			sub := models.Category{
				ID:        uuid.New(),
				ParentID:  &parentID,
				Name:      sc.Name,
				NameEN:    strPtr(sc.NameEN),
				Slug:      sc.Slug,
				SortOrder: (i+1)*10 + j + 1,
			}
			if err := db.Create(&sub).Error; err != nil {
				log.Printf("Failed to create subcategory %s: %v", sc.Name, err)
			}
		}
	}

	log.Println("✅ Categories and subcategories created")
	return nil
}

func strPtr(s string) *string {
	return &s
}
