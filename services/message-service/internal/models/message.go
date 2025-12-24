package models

import (
	"time"

	"github.com/google/uuid"
)

// Message representa uma mensagem no sistema
type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       *uuid.UUID `json:"sender_id"`       // User ID (agente)
	ContactID      *uuid.UUID `json:"contact_id"`      // Contact ID (cliente)
	Content        string     `json:"content"`
	MessageType    string     `json:"message_type"`    // incoming, outgoing, activity
	ContentType    string     `json:"content_type"`    // text, image, file, etc
	Private        bool       `json:"private"`         // Nota interna
	Status         string     `json:"status"`          // sent, delivered, read, failed
	SourceID       string     `json:"source_id"`       // ID externo (WhatsApp, etc)
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateMessageRequest representa o payload de criação de mensagem
type CreateMessageRequest struct {
	ConversationID string     `json:"conversation_id" validate:"required"`
	SenderID       *uuid.UUID `json:"sender_id"`
	ContactID      *uuid.UUID `json:"contact_id"`
	Content        string     `json:"content" validate:"required"`
	MessageType    string     `json:"message_type"`
	ContentType    string     `json:"content_type"`
	Private        bool       `json:"private"`
	SourceID       string     `json:"source_id"`
}

// MessageListResponse representa a resposta de listagem
type MessageListResponse struct {
	Messages []Message `json:"messages"`
	Count    int       `json:"count"`
}
