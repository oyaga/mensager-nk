package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nakamura/chatwoot-go/internal/models"
	"github.com/nakamura/chatwoot-go/internal/websocket"
	"gorm.io/gorm"
)

type IncomingWebhookHandler struct {
	db    *gorm.DB
	wsHub *websocket.Hub
}

func NewIncomingWebhookHandler(db *gorm.DB, wsHub *websocket.Hub) *IncomingWebhookHandler {
	return &IncomingWebhookHandler{db: db, wsHub: wsHub}
}

// HandleIncoming processes incoming webhooks from external services
// Authentication is via API token in header
func (h *IncomingWebhookHandler) HandleIncoming(c *gin.Context) {
	// Get API token from header
	apiToken := c.GetHeader("X-Api-Token")
	if apiToken == "" {
		apiToken = c.GetHeader("api_access_token")
	}
	if apiToken == "" {
		apiToken = c.Query("api_token")
	}

	if apiToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
		return
	}

	// Validate token
	var accessToken models.AccessToken
	if err := h.db.Where("token = ?", apiToken).Preload("User").First(&accessToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API token"})
		return
	}

	// Check expiration
	if accessToken.ExpiresAt != nil && accessToken.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API token expired"})
		return
	}

	// Parse incoming payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get event type
	eventType := c.GetHeader("X-Event-Type")
	if eventType == "" {
		if et, ok := payload["event"].(string); ok {
			eventType = et
		} else {
			eventType = "message"
		}
	}

	log.Printf("Incoming webhook: event=%s user=%s", eventType, accessToken.OwnerID)

	// Process based on event type
	switch eventType {
	case "message", "messages.upsert":
		h.handleMessageEvent(c, accessToken, payload)
	case "message_delivered", "message_read":
		h.handleStatusEvent(c, accessToken, payload)
	case "connection", "qrcode":
		h.handleConnectionEvent(c, accessToken, payload)
	default:
		// Store as generic event for processing
		log.Printf("Unhandled event type: %s", eventType)
		c.JSON(http.StatusOK, gin.H{"status": "received", "event": eventType})
	}
}

// handleMessageEvent processes incoming message webhooks
func (h *IncomingWebhookHandler) handleMessageEvent(c *gin.Context, token models.AccessToken, payload map[string]interface{}) {
	// Extract message data - adapt based on webhook format
	// This supports common formats from Evolution-Go, Chatwoot, etc.

	// Try to extract phone number / contact info
	var phoneNumber string
	var contactName string
	var content string
	var messageType string = "text"
	var attachments []map[string]interface{}

	// Evolution-Go format
	if data, ok := payload["data"].(map[string]interface{}); ok {
		if key, ok := data["key"].(map[string]interface{}); ok {
			if remoteJid, ok := key["remoteJid"].(string); ok {
				phoneNumber = extractPhoneFromJid(remoteJid)
			}
		}
		if message, ok := data["message"].(map[string]interface{}); ok {
			if conv, ok := message["conversation"].(string); ok {
				content = conv
			}
			if extMsg, ok := message["extendedTextMessage"].(map[string]interface{}); ok {
				if text, ok := extMsg["text"].(string); ok {
					content = text
				}
			}
			if img, ok := message["imageMessage"].(map[string]interface{}); ok {
				messageType = "image"
				if url, ok := img["url"].(string); ok {
					attachments = append(attachments, map[string]interface{}{
						"type": "image",
						"url":  url,
					})
				}
				if caption, ok := img["caption"].(string); ok {
					content = caption
				}
			}
			if audio, ok := message["audioMessage"].(map[string]interface{}); ok {
				messageType = "audio"
				if url, ok := audio["url"].(string); ok {
					attachments = append(attachments, map[string]interface{}{
						"type": "audio",
						"url":  url,
					})
				}
			}
			if doc, ok := message["documentMessage"].(map[string]interface{}); ok {
				messageType = "file"
				if url, ok := doc["url"].(string); ok {
					attachments = append(attachments, map[string]interface{}{
						"type":     "file",
						"url":      url,
						"filename": doc["fileName"],
					})
				}
			}
		}
		if pushName, ok := data["pushName"].(string); ok {
			contactName = pushName
		}
	}

	// Generic format fallback
	if phoneNumber == "" {
		if phone, ok := payload["phone"].(string); ok {
			phoneNumber = phone
		} else if from, ok := payload["from"].(string); ok {
			phoneNumber = from
		}
	}
	if content == "" {
		if msg, ok := payload["message"].(string); ok {
			content = msg
		} else if text, ok := payload["text"].(string); ok {
			content = text
		} else if body, ok := payload["body"].(string); ok {
			content = body
		}
	}
	if contactName == "" {
		if name, ok := payload["name"].(string); ok {
			contactName = name
		} else if senderName, ok := payload["sender_name"].(string); ok {
			contactName = senderName
		}
	}

	if phoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number not found in payload"})
		return
	}

	// Identify Inbox and Account ID from URL path (wildcard)
	pathParam := c.Param("pathParam")
	pathParam = strings.TrimPrefix(pathParam, "/")
	segments := strings.Split(pathParam, "/")

	var cleanSegments []string
	for _, s := range segments {
		if s != "" {
			cleanSegments = append(cleanSegments, s)
		}
	}

	var instanceName string
	var accountIDParam string

	if len(cleanSegments) == 1 {
		instanceName = cleanSegments[0]
	} else if len(cleanSegments) >= 2 {
		accountIDParam = cleanSegments[0] // Account ID explícito na URL
		instanceName = cleanSegments[1]
	}

	if instanceName == "" {
		if inst, ok := payload["instance"].(string); ok {
			instanceName = inst
		} else {
			instanceName = "Default WhatsApp"
		}
	}

	// Find or Create Inbox
	var inbox models.Inbox
	// ... User fetch logic follows ...

	// Determine Account ID
	var accountID uuid.UUID
	// accountIDParam já foi extraído acima

	var user models.User
	if err := h.db.Preload("Accounts").First(&user, token.OwnerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if accountIDParam != "" {
		// Use explicit account ID from URL
		// First validate format
		parsedID, err := uuid.Parse(accountIDParam)
		if err != nil {
			// If not a UUID, maybe it's a numeric ID (legacy Chatwoot)?
			// But our models use UUID. Assuming UUID for now as per Go implementation.
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Account ID format (must be UUID)"})
			return
		}

		// Verify user belongs to this account
		hasAccess := false
		for _, acc := range user.Accounts {
			if acc.ID == parsedID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "User does not have access to the specified account"})
			return
		}
		accountID = parsedID
	} else {
		// Fallback: Use user's first account
		if len(user.Accounts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User has no accounts"})
			return
		}
		accountID = user.Accounts[0].ID
	}

	if err := h.db.Where("account_id = ? AND name = ?", accountID, instanceName).First(&inbox).Error; err != nil {
		inbox = models.Inbox{
			AccountID:   accountID,
			Name:        instanceName,
			ChannelType: "whatsapp",
		}
		if err := h.db.Create(&inbox).Error; err != nil {
			log.Printf("Failed to create inbox: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inbox"})
			return
		}
	}

	// Find or create contact
	var contact models.Contact
	if err := h.db.Where("account_id = ? AND phone_number = ?", accountID, phoneNumber).First(&contact).Error; err != nil {
		// Create new contact
		contact = models.Contact{
			Name:        contactName,
			PhoneNumber: phoneNumber,
			AccountID:   accountID,
		}
		if contact.Name == "" {
			contact.Name = phoneNumber
		}
		if err := h.db.Create(&contact).Error; err != nil {
			log.Printf("Failed to create contact: %v", err)
			return
		}
	}

	// Find or create conversation
	var conversation models.Conversation
	if err := h.db.Where("inbox_id = ? AND contact_id = ? AND status = ?", inbox.ID, contact.ID, "open").First(&conversation).Error; err != nil {
		// Create new conversation
		conversation = models.Conversation{
			InboxID:        inbox.ID,
			ContactID:      contact.ID,
			AccountID:      accountID,
			Status:         "open",
			LastActivityAt: time.Now(),
		}
		if err := h.db.Create(&conversation).Error; err != nil {
			log.Printf("Failed to create conversation: %v", err)
			return
		}
	}

	// Create message
	message := models.Message{
		ConversationID: conversation.ID,
		ContactID:      &contact.ID,
		Content:        content,
		ContentType:    messageType,
		MessageType:    "incoming",
		Status:         "delivered",
	}
	h.db.Create(&message)

	// Update conversation
	conversation.LastActivityAt = time.Now()
	h.db.Save(&conversation)

	// Broadcast via WebSocket
	if h.wsHub != nil {
		// Broadcast to conversation room (for users viewing this conversation)
		h.wsHub.BroadcastToRoom(
			conversation.ID.String(),
			"message.created",
			message,
		)

		// Broadcast to global notifications channel (for notification badges)
		h.wsHub.BroadcastToRoom(
			"notifications",
			"message.created",
			map[string]interface{}{
				"conversation_id": conversation.ID,
				"contact_name":    contact.Name,
				"content":         content,
				"inbox_id":        inbox.ID,
			},
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "received",
		"message_id":      message.ID,
		"conversation_id": conversation.ID,
	})
}

// handleStatusEvent processes message status updates
func (h *IncomingWebhookHandler) handleStatusEvent(c *gin.Context, token models.AccessToken, payload map[string]interface{}) {
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// handleConnectionEvent processes connection/QR code events
func (h *IncomingWebhookHandler) handleConnectionEvent(c *gin.Context, token models.AccessToken, payload map[string]interface{}) {
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// extractPhoneFromJid extracts phone number from WhatsApp JID
func extractPhoneFromJid(jid string) string {
	// Format: 5511999999999@s.whatsapp.net
	if len(jid) > 0 {
		atIndex := -1
		for i, c := range jid {
			if c == '@' {
				atIndex = i
				break
			}
		}
		if atIndex > 0 {
			return jid[:atIndex]
		}
	}
	return jid
}

// GenerateToken creates a new API access token for a user
func (h *IncomingWebhookHandler) GenerateToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	c.ShouldBindJSON(&input)

	if input.Name == "" {
		input.Name = "API Token"
	}

	// UserID from middleware is string
	userUUIDStr := userID.(string)
	userUUID, err := uuid.Parse(userUUIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	accessToken := models.AccessToken{
		OwnerID:   userUUID,
		OwnerType: "User",
		Name:      input.Name,
	}

	if err := h.db.Create(&accessToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, accessToken)
}

// ListTokens lists all API tokens for the current user
func (h *IncomingWebhookHandler) ListTokens(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var tokens []models.AccessToken
	h.db.Where("owner_id = ? AND owner_type = ?", userID, "User").Find(&tokens)

	c.JSON(http.StatusOK, tokens)
}

// DeleteToken deletes an API token
func (h *IncomingWebhookHandler) DeleteToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tokenID := c.Param("id")

	result := h.db.Where("id = ? AND owner_id = ?", tokenID, userID).Delete(&models.AccessToken{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token deleted"})
}
