package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONB type for PostgreSQL jsonb columns
type JSONB map[string]interface{}

// Value implements driver.Valuer
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Base model with common fields
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}

// Account represents a Nakamura account/workspace
type Account struct {
	BaseModel
	Name             string `gorm:"not null" json:"name"`
	Status           string `gorm:"default:'active'" json:"status"` // active, suspended
	Locale           string `gorm:"default:'en'" json:"locale"`
	Domain           string `gorm:"uniqueIndex" json:"domain"`
	SupportEmail     string `json:"support_email"`
	AutoResolveTime  int    `gorm:"default:40" json:"auto_resolve_time"` // hours
	FeatureFlags     JSONB  `gorm:"type:jsonb" json:"feature_flags"`
	CustomAttributes JSONB  `gorm:"type:jsonb" json:"custom_attributes"`

	// Relationships
	Users         []User         `gorm:"many2many:account_users;" json:"users,omitempty"`
	Inboxes       []Inbox        `json:"inboxes,omitempty"`
	Conversations []Conversation `json:"conversations,omitempty"`
	Contacts      []Contact      `json:"contacts,omitempty"`
	Teams         []Team         `json:"teams,omitempty"`
}

// User represents a user in the system
type User struct {
	BaseModel
	Name             string `gorm:"not null" json:"name"`
	Email            string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash     string `gorm:"not null" json:"-"`
	DisplayName      string `json:"display_name"`
	Avatar           string `json:"avatar"`
	Role             string `gorm:"default:'agent'" json:"role"`     // administrator, agent, supervisor
	AccessToken      string `gorm:"uniqueIndex" json:"access_token"` // Token único do usuário (estilo Chatwoot original)
	CustomAttributes JSONB  `gorm:"type:jsonb" json:"custom_attributes"`
	Availability     string `gorm:"default:'online'" json:"availability"` // online, busy, offline
	UISettings       JSONB  `gorm:"type:jsonb" json:"ui_settings"`

	// Relationships
	Accounts              []Account      `gorm:"many2many:account_users;" json:"accounts,omitempty"`
	AssignedConversations []Conversation `gorm:"foreignKey:AssigneeID" json:"assigned_conversations,omitempty"`
	Messages              []Message      `gorm:"foreignKey:SenderID" json:"messages,omitempty"`
}

// BeforeCreate hook for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if err := u.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}
	if u.AccessToken == "" {
		u.AccessToken = uuid.New().String()
	}
	return nil
}

// AccountUser is the join table for Account and User
type AccountUser struct {
	AccountID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role      string    `gorm:"default:'agent'"` // administrator, agent
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Inbox represents a communication channel
type Inbox struct {
	BaseModel
	AccountID                  uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
	Name                       string    `gorm:"not null" json:"name"`
	ChannelType                string    `gorm:"not null" json:"channel_type"` // web, whatsapp, email, api, telegram, etc
	ChannelID                  uuid.UUID `gorm:"type:uuid;index" json:"channel_id"`
	AvatarURL                  string    `json:"avatar_url"`
	EnableAutoAssignment       bool      `gorm:"default:true" json:"enable_auto_assignment"`
	GreetingEnabled            bool      `gorm:"default:false" json:"greeting_enabled"`
	GreetingMessage            string    `json:"greeting_message"`
	WorkingHoursEnabled        bool      `gorm:"default:false" json:"working_hours_enabled"`
	OutOfOfficeMessage         string    `json:"out_of_office_message"`
	Timezone                   string    `gorm:"default:'UTC'" json:"timezone"`
	AllowMessagesAfterResolved bool      `gorm:"default:true" json:"allow_messages_after_resolved"`

	// Relationships
	Account       Account        `json:"account,omitempty"`
	Conversations []Conversation `json:"conversations,omitempty"`
	Contacts      []Contact      `gorm:"many2many:inbox_contacts;" json:"contacts,omitempty"`
}

// Contact represents a customer/contact
type Contact struct {
	BaseModel
	AccountID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"account_id"`
	Name                 string     `json:"name"`
	Email                string     `gorm:"index" json:"email"`
	PhoneNumber          string     `gorm:"index" json:"phone_number"`
	Avatar               string     `json:"avatar"`
	Identifier           string     `gorm:"index" json:"identifier"` // External ID
	CustomAttributes     JSONB      `gorm:"type:jsonb" json:"custom_attributes"`
	AdditionalAttributes JSONB      `gorm:"type:jsonb" json:"additional_attributes"`
	LastActivityAt       *time.Time `json:"last_activity_at"`

	// Relationships
	Account       Account        `json:"account,omitempty"`
	Conversations []Conversation `json:"conversations,omitempty"`
	Inboxes       []Inbox        `gorm:"many2many:inbox_contacts;" json:"inboxes,omitempty"`
}

// Conversation represents a conversation thread
type Conversation struct {
	BaseModel
	AccountID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"account_id"`
	InboxID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"inbox_id"`
	ContactID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"contact_id"`
	AssigneeID *uuid.UUID `gorm:"type:uuid;index" json:"assignee_id"`
	TeamID     *uuid.UUID `gorm:"type:uuid;index" json:"team_id"`
	Status     string     `gorm:"default:'open';index" json:"status"` // open, resolved, pending, snoozed
	// Priority             string         `gorm:"default:'none'" json:"priority"`     // urgent, high, medium, low, none
	DisplayID            int        `gorm:"autoIncrement" json:"display_id"`
	AdditionalAttributes JSONB      `gorm:"type:jsonb" json:"additional_attributes"`
	CustomAttributes     JSONB      `gorm:"type:jsonb" json:"custom_attributes"`
	LastActivityAt       time.Time  `gorm:"index" json:"last_activity_at"`
	FirstReplyCreatedAt  *time.Time `json:"first_reply_created_at"`
	AgentLastSeenAt      *time.Time `json:"agent_last_seen_at"`
	ContactLastSeenAt    *time.Time `json:"contact_last_seen_at"`
	SnoozedUntil         *time.Time `json:"snoozed_until"`

	// Relationships
	Account  Account   `json:"account,omitempty"`
	Inbox    Inbox     `json:"inbox,omitempty"`
	Contact  Contact   `json:"contact,omitempty"`
	Assignee *User     `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Team     *Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Labels   []Label   `gorm:"many2many:conversation_labels;" json:"labels,omitempty"`
}

// Message represents a message in a conversation
type Message struct {
	BaseModel
	ConversationID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID          *uuid.UUID `gorm:"type:uuid;index" json:"sender_id"`       // User ID if sent by agent
	ContactID         *uuid.UUID `gorm:"type:uuid;index" json:"contact_id"`      // Contact ID if sent by customer
	MessageType       string     `gorm:"default:'incoming'" json:"message_type"` // incoming, outgoing, activity, template
	ContentType       string     `gorm:"default:'text'" json:"content_type"`     // text, input_select, cards, form, article, etc
	Content           string     `gorm:"type:text" json:"content"`
	Private           bool       `gorm:"default:false" json:"private"` // Internal note
	Status            string     `gorm:"default:'sent'" json:"status"` // sent, delivered, read, failed
	SourceID          string     `gorm:"index" json:"source_id"`       // External message ID
	ContentAttributes JSONB      `gorm:"type:jsonb" json:"content_attributes"`
	ExternalSourceID  string     `json:"external_source_id"`

	// Relationships
	Conversation Conversation `json:"conversation,omitempty"`
	Sender       *User        `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Contact      *Contact     `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	Attachments  []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a file attachment
type Attachment struct {
	BaseModel
	MessageID uuid.UUID `gorm:"type:uuid;not null;index" json:"message_id"`
	FileType  string    `gorm:"not null" json:"file_type"` // image, audio, video, file
	FileURL   string    `gorm:"not null" json:"file_url"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`

	// Relationships
	Message Message `json:"message,omitempty"`
}

// Team represents a team of agents
type Team struct {
	BaseModel
	AccountID       uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
	Name            string    `gorm:"not null" json:"name"`
	Description     string    `json:"description"`
	AllowAutoAssign bool      `gorm:"default:true" json:"allow_auto_assign"`

	// Relationships
	Account       Account        `json:"account,omitempty"`
	Members       []User         `gorm:"many2many:team_members;" json:"members,omitempty"`
	Conversations []Conversation `json:"conversations,omitempty"`
}

// Label represents a conversation label/tag
type Label struct {
	BaseModel
	AccountID     uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
	Title         string    `gorm:"not null" json:"title"`
	Description   string    `json:"description"`
	Color         string    `gorm:"default:'#1f93ff'" json:"color"`
	ShowOnSidebar bool      `gorm:"default:false" json:"show_on_sidebar"`

	// Relationships
	Account       Account        `json:"account,omitempty"`
	Conversations []Conversation `gorm:"many2many:conversation_labels;" json:"conversations,omitempty"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	BaseModel
	AccountID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"account_id"`
	InboxID       *uuid.UUID `gorm:"type:uuid;index" json:"inbox_id"`
	Name          string     `json:"name"`
	URL           string     `gorm:"not null" json:"url"`
	WebhookType   string     `gorm:"default:'account'" json:"webhook_type"` // account, inbox
	Subscriptions []string   `gorm:"type:text[]" json:"subscriptions"`      // conversation_created, message_created, etc

	// Relationships
	Account Account `json:"account,omitempty"`
	Inbox   *Inbox  `gorm:"foreignKey:InboxID" json:"inbox,omitempty"`
}

// AccessToken represents an API access token for users or platform apps
type AccessToken struct {
	BaseModel
	OwnerID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_id"`
	OwnerType string     `gorm:"not null" json:"owner_type"` // User, PlatformApp
	Token     string     `gorm:"uniqueIndex;not null" json:"token"`
	Name      string     `json:"name"`       // Token name/description
	ExpiresAt *time.Time `json:"expires_at"` // Optional expiration

	// Relationships
	User *User `gorm:"foreignKey:OwnerID" json:"user,omitempty"`
}

// BeforeCreate hook to generate token
func (at *AccessToken) BeforeCreate(tx *gorm.DB) error {
	if at.Token == "" {
		at.Token = "v1_" + uuid.New().String()
	}
	return nil
}
