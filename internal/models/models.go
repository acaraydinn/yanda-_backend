package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// User represents a platform user (customer or yandaş)
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        *string        `gorm:"uniqueIndex;size:255" json:"email,omitempty"`
	Phone        *string        `gorm:"uniqueIndex;size:20" json:"phone,omitempty"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	FullName     string         `gorm:"size:255;not null" json:"full_name"`
	AvatarURL    *string        `gorm:"type:text" json:"avatar_url,omitempty"`
	Role         string         `gorm:"size:20;default:customer" json:"role"` // customer, yandas, admin
	IsVerified   bool           `gorm:"default:false" json:"is_verified"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	YandasProfile *YandasProfile `gorm:"foreignKey:UserID" json:"yandas_profile,omitempty"`
	DeviceTokens  []DeviceToken  `gorm:"foreignKey:UserID" json:"-"`
	Subscription  *Subscription  `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
}

// YandasProfile contains extended data for Yandaş users
type YandasProfile struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	Bio               *string   `gorm:"type:text" json:"bio,omitempty"`
	InstagramHandle   *string   `gorm:"size:100" json:"instagram_handle,omitempty"`
	InstagramVerified bool      `gorm:"default:false" json:"instagram_verified"`
	// Document URLs - Front and Back for ID and License
	KimlikOnURL     *string `gorm:"type:text" json:"-"` // ID Card Front
	KimlikArkaURL   *string `gorm:"type:text" json:"-"` // ID Card Back
	EhliyetOnURL    *string `gorm:"type:text" json:"-"` // Driver's License Front
	EhliyetArkaURL  *string `gorm:"type:text" json:"-"` // Driver's License Back
	AdliSicilPDFURL *string `gorm:"type:text" json:"-"` // Criminal Record PDF
	// Verification status for each document
	KimlikOnVerified    bool           `gorm:"default:false" json:"kimlik_on_verified"`
	KimlikArkaVerified  bool           `gorm:"default:false" json:"kimlik_arka_verified"`
	EhliyetOnVerified   bool           `gorm:"default:false" json:"ehliyet_on_verified"`
	EhliyetArkaVerified bool           `gorm:"default:false" json:"ehliyet_arka_verified"`
	AdliSicilVerified   bool           `gorm:"default:false" json:"adli_sicil_verified"`
	ApprovalStatus      string         `gorm:"size:20;default:pending" json:"approval_status"` // pending, approved, rejected
	ApprovedBy          *uuid.UUID     `gorm:"type:uuid" json:"-"`
	ApprovedAt          *time.Time     `json:"approved_at,omitempty"`
	RejectionReason     *string        `gorm:"type:text" json:"rejection_reason,omitempty"`
	RatingAvg           float64        `gorm:"type:decimal(3,2);default:0" json:"rating_avg"`
	TotalJobs           int            `gorm:"default:0" json:"total_jobs"`
	IsAvailable         bool           `gorm:"default:false" json:"is_available"`
	Latitude            *float64       `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`
	Longitude           *float64       `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`
	ServiceCities       pq.StringArray `gorm:"type:text[]" json:"service_cities"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User     User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Services []YandasService `gorm:"foreignKey:YandasID" json:"services,omitempty"`
}

// Category represents service categories
type Category struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ParentID      *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Name          string     `gorm:"size:100;not null" json:"name"`
	NameEN        *string    `gorm:"size:100" json:"name_en,omitempty"`
	Slug          string     `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Icon          *string    `gorm:"size:50" json:"icon,omitempty"`
	Description   *string    `gorm:"type:text" json:"description,omitempty"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	SortOrder     int        `gorm:"default:0" json:"sort_order"`
	SubCategories []Category `gorm:"foreignKey:ParentID" json:"sub_categories,omitempty"`
}

// YandasService represents a service/package offered by a Yandaş
type YandasService struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	YandasID        uuid.UUID      `gorm:"type:uuid;not null" json:"yandas_id"`
	CategoryID      uuid.UUID      `gorm:"type:uuid" json:"category_id"`
	Title           string         `gorm:"size:255;not null" json:"title"`
	Description     *string        `gorm:"type:text" json:"description,omitempty"`
	BasePrice       float64        `gorm:"type:decimal(10,2);not null" json:"base_price"`
	Currency        string         `gorm:"size:3;default:TRY" json:"currency"`
	DurationMinutes *int           `json:"duration_minutes,omitempty"`
	Includes        pq.StringArray `gorm:"type:text[]" json:"includes,omitempty"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// Order represents a booking/order
type Order struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderNumber        string         `gorm:"size:20;uniqueIndex;not null" json:"order_number"`
	CustomerID         uuid.UUID      `gorm:"type:uuid;not null" json:"customer_id"`
	YandasID           uuid.UUID      `gorm:"type:uuid;not null" json:"yandas_id"`
	ServiceID          uuid.UUID      `gorm:"type:uuid" json:"service_id"`
	Status             string         `gorm:"size:30;default:pending" json:"status"` // pending, accepted, in_progress, completed, cancelled, disputed
	AgreedPrice        float64        `gorm:"type:decimal(10,2);not null" json:"agreed_price"`
	Currency           string         `gorm:"size:3;default:TRY" json:"currency"`
	LocationAddress    *string        `gorm:"type:text" json:"location_address,omitempty"`
	Latitude           *float64       `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`
	Longitude          *float64       `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`
	ScheduledAt        *time.Time     `json:"scheduled_at,omitempty"`
	StartedAt          *time.Time     `json:"started_at,omitempty"`
	CompletedAt        *time.Time     `json:"completed_at,omitempty"`
	CustomerNotes      *string        `gorm:"type:text" json:"customer_notes,omitempty"`
	YandasNotes        *string        `gorm:"type:text" json:"yandas_notes,omitempty"`
	CancellationReason *string        `gorm:"type:text" json:"cancellation_reason,omitempty"`
	CancelledBy        *uuid.UUID     `gorm:"type:uuid" json:"cancelled_by,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Customer *User          `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Yandas   *YandasProfile `gorm:"foreignKey:YandasID" json:"yandas,omitempty"`
	Service  *YandasService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Review   *Review        `gorm:"foreignKey:OrderID" json:"review,omitempty"`
}

// Review represents a rating/review for an order
type Review struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"order_id"`
	ReviewerID  uuid.UUID `gorm:"type:uuid;not null" json:"reviewer_id"`
	RevieweeID  uuid.UUID `gorm:"type:uuid;not null" json:"reviewee_id"`
	Rating      int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment     *string   `gorm:"type:text" json:"comment,omitempty"`
	IsAnonymous bool      `gorm:"default:false" json:"is_anonymous"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Reviewer *User `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID       *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
	CustomerID    uuid.UUID  `gorm:"type:uuid;not null" json:"customer_id"`
	YandasID      uuid.UUID  `gorm:"type:uuid;not null" json:"yandas_id"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Customer *User     `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Yandas   *User     `gorm:"foreignKey:YandasID" json:"yandas,omitempty"`
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// Message represents a chat message
type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null" json:"conversation_id"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	MessageType    string    `gorm:"size:20;default:text" json:"message_type"` // text, image, location, system
	IsRead         bool      `gorm:"default:false" json:"is_read"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// Subscription represents a premium subscription
type Subscription struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID                 uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	PlanType               string     `gorm:"size:20;not null" json:"plan_type"`    // monthly, yearly
	Status                 string     `gorm:"size:20;default:active" json:"status"` // active, cancelled, expired
	Provider               string     `gorm:"size:20;not null" json:"provider"`     // revenuecat, stripe
	ProviderSubscriptionID *string    `gorm:"size:255" json:"-"`
	CurrentPeriodStart     *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end,omitempty"`
	CancelledAt            *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt              time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// DeviceToken represents a push notification token
type DeviceToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Token     string    `gorm:"type:text;not null" json:"token"`
	Platform  string    `gorm:"size:10;not null" json:"platform"` // ios, android, web
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// AuditLog represents admin action logs
type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AdminID    uuid.UUID  `gorm:"type:uuid;not null" json:"admin_id"`
	Action     string     `gorm:"size:100;not null" json:"action"`
	EntityType *string    `gorm:"size:50" json:"entity_type,omitempty"`
	EntityID   *uuid.UUID `gorm:"type:uuid" json:"entity_id,omitempty"`
	OldValues  *string    `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues  *string    `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress  *string    `gorm:"size:45" json:"ip_address,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Admin *User `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

// Notification represents in-app notifications
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	Type      string    `gorm:"size:50" json:"type"` // order, chat, system, promotion
	Data      *string   `gorm:"type:jsonb" json:"data,omitempty"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// SupportTicket represents a support request
type SupportTicket struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	AssignedTo  *uuid.UUID `gorm:"type:uuid" json:"assigned_to,omitempty"`
	Subject     string     `gorm:"size:255;not null" json:"subject"`
	Description string     `gorm:"type:text;not null" json:"description"`
	Category    string     `gorm:"size:50;default:general" json:"category"` // general, order, payment, account, technical
	Priority    string     `gorm:"size:20;default:normal" json:"priority"`  // low, normal, high, urgent
	Status      string     `gorm:"size:20;default:open" json:"status"`      // open, pending, in_progress, resolved, closed
	OrderID     *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User     *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Assignee *User            `gorm:"foreignKey:AssignedTo" json:"assignee,omitempty"`
	Messages []SupportMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
}

// SupportMessage represents a message in a support conversation
type SupportMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TicketID  uuid.UUID `gorm:"type:uuid;not null" json:"ticket_id"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// Favorite represents a user's bookmarked Yandaş
type Favorite struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	YandasID  uuid.UUID `gorm:"type:uuid;not null" json:"yandas_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User   *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Yandas *YandasProfile `gorm:"foreignKey:YandasID" json:"yandas,omitempty"`
}

// CallLog represents a voice/video call record
type CallLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CallerID   uuid.UUID  `gorm:"type:uuid;not null" json:"caller_id"`
	CalleeID   uuid.UUID  `gorm:"type:uuid;not null" json:"callee_id"`
	OrderID    *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
	CallType   string     `gorm:"size:20;not null" json:"call_type"` // voice, video
	Status     string     `gorm:"size:20;not null" json:"status"`    // initiated, ringing, answered, ended, missed, declined
	Duration   int        `gorm:"default:0" json:"duration"`         // seconds
	ChannelID  *string    `gorm:"size:255" json:"channel_id,omitempty"`
	StartedAt  time.Time  `gorm:"autoCreateTime" json:"started_at"`
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`

	// Relations
	Caller *User `gorm:"foreignKey:CallerID" json:"caller,omitempty"`
	Callee *User `gorm:"foreignKey:CalleeID" json:"callee,omitempty"`
}
