package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yandas/backend/internal/config"
)

type LegalHandler struct {
	cfg *config.Config
}

func NewLegalHandler(cfg *config.Config) *LegalHandler {
	return &LegalHandler{cfg: cfg}
}

func (h *LegalHandler) PrivacyPolicy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "Gizlilik Politikası",
		"content": `
# Gizlilik Politikası

Son güncelleme: Şubat 2026

## 1. Giriş
Yandaş ("biz", "bizim") olarak gizliliğinize saygı duyuyoruz.

## 2. Toplanan Veriler
- Kimlik bilgileri (ad, e-posta, telefon)
- Konum verileri (hizmet sağlanması için)
- Ödeme bilgileri (işlem güvenliği için)

## 3. Verilerin Kullanımı
- Hizmet sağlanması
- İletişim
- Güvenlik

## 4. Veri Güvenliği
Verileriniz şifrelenerek saklanır.

## 5. Haklarınız
KVKK kapsamında verilerinize erişim, düzeltme ve silme hakkına sahipsiniz.

## 6. İletişim
privacy@yandas.app
`,
	})
}

func (h *LegalHandler) TermsOfService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "Kullanım Koşulları",
		"content": `
# Kullanım Koşulları

Son güncelleme: Şubat 2026

## 1. Kabul
Bu uygulamayı kullanarak bu koşulları kabul etmiş olursunuz.

## 2. Hizmet Tanımı
Yandaş, vekalet ve yerinden hizmet platformudur.

## 3. Kullanıcı Sorumlulukları
- Doğru bilgi sağlamak
- Yasalara uymak
- Diğer kullanıcılara saygı göstermek

## 4. Yandaş Sorumlulukları
- Onaylı belgeler sunmak
- Profesyonel hizmet vermek
- Müşteri güvenliğini sağlamak

## 5. Ödeme
Ödemeler platform dışında yapılır.

## 6. İletişim
legal@yandas.app
`,
	})
}

func (h *LegalHandler) KVKK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "KVKK Aydınlatma Metni",
		"content": `
# KVKK Aydınlatma Metni

6698 sayılı Kişisel Verilerin Korunması Kanunu kapsamında:

## Veri Sorumlusu
Yandaş Teknoloji A.Ş.

## Toplanan Veriler
- Kimlik bilgileri
- İletişim bilgileri
- Konum verileri

## İşleme Amaçları
- Sözleşmenin ifası
- Yasal yükümlülükler

## Haklarınız
- Bilgi alma
- Erişim
- Düzeltme
- Silme
- İtiraz

## İletişim
kvkk@yandas.app
`,
	})
}
