package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nakamura/chatwoot-go/services/message-service/internal/database"
	"github.com/nakamura/chatwoot-go/services/message-service/internal/models"
)

type MessageHandler struct {
	db *database.DB
}

func NewMessageHandler(db *database.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

// CreateMessage cria uma nova mensagem
func (h *MessageHandler) CreateMessage(c *fiber.Ctx) error {
	var req models.CreateMessageRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validações básicas
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid conversation_id",
		})
	}

	// Defaults
	if req.MessageType == "" {
		req.MessageType = "incoming"
	}
	if req.ContentType == "" {
		req.ContentType = "text"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert na tabela messages
	query := `
		INSERT INTO messages (
			id, conversation_id, sender_id, contact_id, content,
			message_type, content_type, private, status, source_id,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		) RETURNING id, conversation_id, sender_id, contact_id, content,
		           message_type, content_type, private, status, source_id,
		           created_at, updated_at
	`

	var message models.Message
	err = h.db.Pool.QueryRow(
		ctx, query,
		conversationID,
		req.SenderID,
		req.ContactID,
		req.Content,
		req.MessageType,
		req.ContentType,
		req.Private,
		"sent",
		req.SourceID,
	).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.ContactID,
		&message.Content,
		&message.MessageType,
		&message.ContentType,
		&message.Private,
		&message.Status,
		&message.SourceID,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create message",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(message)
}

// GetMessages lista mensagens de uma conversa
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	conversationID := c.Params("conversation_id")

	if _, err := uuid.Parse(conversationID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid conversation_id",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, conversation_id, sender_id, contact_id, content,
		       message_type, content_type, private, status, source_id,
		       created_at, updated_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := h.db.Pool.Query(ctx, query, conversationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch messages",
		})
	}
	defer rows.Close()

	messages := make([]models.Message, 0)

	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.ContactID,
			&msg.Content,
			&msg.MessageType,
			&msg.ContentType,
			&msg.Private,
			&msg.Status,
			&msg.SourceID,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return c.JSON(models.MessageListResponse{
		Messages: messages,
		Count:    len(messages),
	})
}

// GetMessage busca uma mensagem específica por ID
func (h *MessageHandler) GetMessage(c *fiber.Ctx) error {
	messageID := c.Params("id")

	if _, err := uuid.Parse(messageID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid message_id",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, conversation_id, sender_id, contact_id, content,
		       message_type, content_type, private, status, source_id,
		       created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	var message models.Message
	err := h.db.Pool.QueryRow(ctx, query, messageID).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.ContactID,
		&message.Content,
		&message.MessageType,
		&message.ContentType,
		&message.Private,
		&message.Status,
		&message.SourceID,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Message not found",
		})
	}

	return c.JSON(message)
}
