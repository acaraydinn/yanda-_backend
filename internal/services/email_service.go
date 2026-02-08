package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/yandas/backend/internal/config"
)

// EmailService handles sending emails via SMTP
type EmailService struct {
	cfg *config.Config
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendOTPEmail sends a beautiful OTP verification email
func (s *EmailService) SendOTPEmail(to, otp, userName string) error {
	if s.cfg.SMTPUser == "" || s.cfg.SMTPPassword == "" {
		log.Printf("[EMAIL FALLBACK] OTP for %s: %s\n", to, otp)
		return nil
	}

	subject := "YANDAŞ - E-posta Doğrulama Kodu"
	body := s.buildOTPEmailHTML(otp, userName)

	return s.sendHTML(to, subject, body)
}

// SendWelcomeEmail sends a welcome email after verification
func (s *EmailService) SendWelcomeEmail(to, userName string) error {
	if s.cfg.SMTPUser == "" {
		return nil
	}

	subject := "YANDAŞ'a Hoş Geldiniz! 🎉"
	body := s.buildWelcomeEmailHTML(userName)

	return s.sendHTML(to, subject, body)
}

func (s *EmailService) sendHTML(to, subject, body string) error {
	from := s.cfg.SMTPFrom
	fromName := s.cfg.SMTPFromName

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	// TLS config - InsecureSkipVerify needed if cert hostname doesn't match SMTP host
	tlsConfig := &tls.Config{
		ServerName:         s.cfg.SMTPHost,
		InsecureSkipVerify: true,
	}

	// Connect to SMTP server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		// Try STARTTLS if direct TLS fails
		log.Printf("Direct TLS failed, trying STARTTLS: %v", err)
		c, dialErr := smtp.Dial(addr)
		if dialErr != nil {
			return fmt.Errorf("SMTP dial error: %w", dialErr)
		}
		defer c.Close()
		if err := c.StartTLS(tlsConfig); err != nil {
			log.Printf("STARTTLS failed, sending plain: %v", err)
		}
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth error: %w", err)
		}
		if err := c.Mail(from); err != nil {
			return fmt.Errorf("SMTP mail error: %w", err)
		}
		if err := c.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP rcpt error: %w", err)
		}
		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("SMTP data error: %w", err)
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return fmt.Errorf("SMTP write error: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("SMTP close error: %w", err)
		}
		c.Quit()
		log.Printf("✅ Email sent (STARTTLS) to: %s\n", to)
		return nil
	}

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("SMTP client error: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth error: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SMTP mail error: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP rcpt error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP data error: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("SMTP write error: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("SMTP close error: %w", err)
	}

	client.Quit()
	log.Printf("✅ Email sent to: %s\n", to)
	return nil
}

func (s *EmailService) buildOTPEmailHTML(otp, userName string) string {
	// Split OTP into individual characters for styled boxes
	otpChars := strings.Split(otp, "")
	otpBoxes := ""
	for _, ch := range otpChars {
		otpBoxes += fmt.Sprintf(`<td style="width:48px;height:56px;text-align:center;font-size:28px;font-weight:700;color:#6C3CE1;background:#F3EFFE;border-radius:12px;border:2px solid #6C3CE1;font-family:'Segoe UI',sans-serif;">%s</td><td style="width:8px;"></td>`, ch)
	}

	if userName == "" {
		userName = "Değerli Kullanıcı"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="tr">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#F5F3FF;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#F5F3FF;padding:40px 0;">
  <tr><td align="center">
    <table width="480" cellpadding="0" cellspacing="0" style="background:#FFFFFF;border-radius:20px;overflow:hidden;box-shadow:0 4px 24px rgba(108,60,225,0.08);">
      <!-- Header -->
      <tr><td style="background:linear-gradient(135deg,#6C3CE1 0%%,#9B6DFF 100%%);padding:32px 40px;text-align:center;">
        <h1 style="margin:0;color:#FFFFFF;font-size:28px;font-weight:800;letter-spacing:-0.5px;">YANDAŞ</h1>
        <p style="margin:8px 0 0;color:rgba(255,255,255,0.85);font-size:14px;">Güvenli Hizmet Platformu</p>
      </td></tr>
      <!-- Body -->
      <tr><td style="padding:40px;">
        <h2 style="margin:0 0 8px;color:#1A1A2E;font-size:22px;font-weight:700;">E-posta Doğrulama</h2>
        <p style="margin:0 0 24px;color:#666;font-size:15px;line-height:1.6;">
          Merhaba <strong>%s</strong>,<br/>
          Hesabını doğrulamak için aşağıdaki kodu uygulamaya gir:
        </p>
        <!-- OTP Code -->
        <table cellpadding="0" cellspacing="0" style="margin:0 auto 24px;">
          <tr>%s</tr>
        </table>
        <p style="margin:0 0 24px;color:#999;font-size:13px;text-align:center;">
          Bu kod <strong>5 dakika</strong> içinde geçerliliğini yitirecek.
        </p>
        <!-- Divider -->
        <hr style="border:none;border-top:1px solid #EEE;margin:24px 0;">
        <!-- Security Notice -->
        <table cellpadding="0" cellspacing="0" width="100%%">
          <tr>
            <td style="width:36px;vertical-align:top;"><div style="width:36px;height:36px;background:#FFF3E0;border-radius:10px;text-align:center;line-height:36px;font-size:18px;">🔒</div></td>
            <td style="padding-left:12px;">
              <p style="margin:0;color:#666;font-size:12px;line-height:1.5;">
                Bu kodu kimseyle paylaşmayın. YANDAŞ ekibi asla doğrulama kodunuzu istemez.
              </p>
            </td>
          </tr>
        </table>
      </td></tr>
      <!-- Footer -->
      <tr><td style="background:#FAFAFA;padding:24px 40px;text-align:center;border-top:1px solid #F0F0F0;">
        <p style="margin:0;color:#AAA;font-size:12px;">
          © 2026 YANDAŞ. Tüm hakları saklıdır.<br/>
          Bu e-postayı siz talep ettiyseniz bir işlem yapmanıza gerek yok.
        </p>
      </td></tr>
    </table>
  </td></tr>
</table>
</body>
</html>`, userName, otpBoxes)
}

func (s *EmailService) buildWelcomeEmailHTML(userName string) string {
	if userName == "" {
		userName = "Değerli Kullanıcı"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="tr">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#F5F3FF;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#F5F3FF;padding:40px 0;">
  <tr><td align="center">
    <table width="480" cellpadding="0" cellspacing="0" style="background:#FFFFFF;border-radius:20px;overflow:hidden;box-shadow:0 4px 24px rgba(108,60,225,0.08);">
      <tr><td style="background:linear-gradient(135deg,#6C3CE1 0%%,#9B6DFF 100%%);padding:40px;text-align:center;">
        <div style="font-size:48px;margin-bottom:16px;">🎉</div>
        <h1 style="margin:0;color:#FFFFFF;font-size:28px;font-weight:800;">Hoş Geldiniz!</h1>
      </td></tr>
      <tr><td style="padding:40px;">
        <p style="margin:0 0 16px;color:#1A1A2E;font-size:16px;line-height:1.6;">
          Merhaba <strong>%s</strong>,
        </p>
        <p style="margin:0 0 24px;color:#666;font-size:15px;line-height:1.6;">
          YANDAŞ ailesine katıldığınız için teşekkür ederiz! Artık güvenilir hizmet sağlayıcılarımızla tanışabilir ve hizmet alabilirsiniz.
        </p>
        <table cellpadding="0" cellspacing="0" width="100%%">
          <tr>
            <td style="padding:12px 0;"><span style="color:#6C3CE1;font-weight:600;">✓</span> <span style="color:#333;">Yandaş'ları keşfedin</span></td>
          </tr>
          <tr>
            <td style="padding:12px 0;"><span style="color:#6C3CE1;font-weight:600;">✓</span> <span style="color:#333;">Güvenle hizmet alın</span></td>
          </tr>
          <tr>
            <td style="padding:12px 0;"><span style="color:#6C3CE1;font-weight:600;">✓</span> <span style="color:#333;">Değerlendirme yapın</span></td>
          </tr>
        </table>
      </td></tr>
      <tr><td style="background:#FAFAFA;padding:24px 40px;text-align:center;border-top:1px solid #F0F0F0;">
        <p style="margin:0;color:#AAA;font-size:12px;">© 2026 YANDAŞ. Tüm hakları saklıdır.</p>
      </td></tr>
    </table>
  </td></tr>
</table>
</body>
</html>`, userName)
}
