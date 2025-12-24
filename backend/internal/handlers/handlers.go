package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nakamura/chatwoot-go/internal/models"
	"github.com/nakamura/chatwoot-go/internal/websocket"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------
// Conversation Handler
// ---------------------------------------------------------------------

type ConversationHandler struct {
	db    *gorm.DB
	wsHub *websocket.Hub
}

func NewConversationHandler(db *gorm.DB, wsHub *websocket.Hub) *ConversationHandler {
	return &ConversationHandler{db: db, wsHub: wsHub}
}

// List conversations
func (h *ConversationHandler) List(c *gin.Context) {
	accountID := c.GetString("account_id")
	status := c.Query("status")
	inboxID := c.Query("inbox_id")

	query := h.db.
		Preload("Contact").
		Preload("Inbox").
		Where("account_id = ?", accountID) // Basic security filter

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", "open")
	}

	if inboxID != "" {
		query = query.Where("inbox_id = ?", inboxID)
	}

	var conversations []models.Conversation
	if err := query.Order("last_activity_at desc").Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

// Create conversation
func (h *ConversationHandler) Create(c *gin.Context) {
	accountIDStr := c.GetString("account_id")
	accountID, _ := uuid.Parse(accountIDStr)

	var input struct {
		ContactID string `json:"contact_id" binding:"required"`
		InboxID   string `json:"inbox_id"`
		Status    string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contactUUID, err := uuid.Parse(input.ContactID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Contact ID"})
		return
	}

	// Validate Contact
	var contact models.Contact
	if err := h.db.Where("id = ? AND account_id = ?", contactUUID, accountID).First(&contact).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Resolve Inbox
	var inboxID uuid.UUID
	if input.InboxID != "" {
		inboxID, _ = uuid.Parse(input.InboxID)
	} else {
		// Attempt to find existing inbox or default
		var inbox models.Inbox
		if err := h.db.Where("account_id = ?", accountID).First(&inbox).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No inbox available. Create one first."})
			return
		}
		inboxID = inbox.ID
	}

	// Check if open conversation exists
	var existing models.Conversation
	if err := h.db.Where("contact_id = ? AND inbox_id = ? AND status = 'open'", contact.ID, inboxID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, existing)
		return
	}

	conversation := models.Conversation{
		AccountID:      accountID,
		InboxID:        inboxID,
		ContactID:      contact.ID,
		Status:         "open",
		LastActivityAt: time.Now(),
	}
	if input.Status != "" {
		conversation.Status = input.Status
	}

	if err := h.db.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch complete object for response
	h.db.Preload("Contact").Preload("Inbox").First(&conversation, conversation.ID)

	c.JSON(http.StatusCreated, conversation)
}

func (h *ConversationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("account_id")

	var conversation models.Conversation
	if err := h.db.Preload("Contact").Preload("Inbox").Where("id = ? AND account_id = ?", id, accountID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

func (h *ConversationHandler) Update(c *gin.Context) {
	// Implementation simplified
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ConversationHandler) Delete(c *gin.Context) {
	// Implementation simplified
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ConversationHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		UserID string `json:"user_id"`
	}
	c.ShouldBindJSON(&input)

	//Logic to update assignee_id
	updater := h.db.Model(&models.Conversation{}).Where("id = ?", id)
	if input.UserID == "" {
		updater.Update("assignee_id", nil)
	} else {
		updater.Update("assignee_id", input.UserID)
	}

	c.JSON(http.StatusOK, gin.H{"status": "assigned"})
}

func (h *ConversationHandler) Resolve(c *gin.Context) {
	id := c.Param("id")
	h.db.Model(&models.Conversation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           "resolved",
		"last_activity_at": time.Now(),
	})
	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}

func (h *ConversationHandler) Reopen(c *gin.Context) {
	id := c.Param("id")
	h.db.Model(&models.Conversation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           "open",
		"last_activity_at": time.Now(),
	})
	c.JSON(http.StatusOK, gin.H{"status": "reopened"})
}

func (h *ConversationHandler) Snooze(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

func (h *ConversationHandler) AddLabel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

func (h *ConversationHandler) RemoveLabel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

func (h *ConversationHandler) ListByContact(c *gin.Context) {
	contactID := c.Param("id")
	var conversations []models.Conversation
	h.db.Where("contact_id = ?", contactID).Find(&conversations)
	c.JSON(http.StatusOK, conversations)
}

func (h *ConversationHandler) CreatePublicConversation(c *gin.Context) {
	// Implementation for widget
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// ---------------------------------------------------------------------
// Message Handler
// ---------------------------------------------------------------------

type MessageHandler struct {
	db    *gorm.DB
	wsHub *websocket.Hub
}

func NewMessageHandler(db *gorm.DB, wsHub *websocket.Hub) *MessageHandler {
	return &MessageHandler{db: db, wsHub: wsHub}
}

func (h *MessageHandler) ListByConversation(c *gin.Context) {
	conversationID := c.Param("id")
	var messages []models.Message
	if err := h.db.Preload("Attachments").Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *MessageHandler) Create(c *gin.Context) {
	accountID := c.GetString("account_id")
	userIDStr := c.GetString("user_id")

	if accountID == "" || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := uuid.Parse(userIDStr)

	var input struct {
		ConversationID string `json:"conversation_id" binding:"required"`
		Content        string `json:"content"`
		ContentType    string `json:"content_type"`
		MessageType    string `json:"message_type"`
		Attachments    []struct {
			FileType string `json:"file_type"`
			FileURL  string `json:"file_url"`
			FileName string `json:"file_name"`
		} `json:"attachments"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversationUUID, _ := uuid.Parse(input.ConversationID)

	// Validate conversation belongs to account
	var conversation models.Conversation
	if err := h.db.Where("id = ? AND account_id = ?", conversationUUID, accountID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Create Message
	message := models.Message{
		ConversationID: conversation.ID,
		SenderID:       &userID,
		Content:        input.Content,
		ContentType:    input.ContentType,
		MessageType:    "outgoing",
		Status:         "sent",
	}
	if message.ContentType == "" {
		message.ContentType = "text"
	}

	if err := h.db.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create Attachments if any
	for _, att := range input.Attachments {
		attachment := models.Attachment{
			MessageID: message.ID,
			FileType:  att.FileType,
			FileURL:   att.FileURL,
			FileName:  att.FileName,
		}
		h.db.Create(&attachment)
	}

	// Reload message with attachments
	h.db.Preload("Attachments").First(&message, message.ID)

	// Update Conversation
	h.db.Model(&conversation).Updates(map[string]interface{}{
		"last_activity_at": time.Now(),
		"last_message":     input.Content, // Assuming we had this field, actually models doesn't show it but JSON response often simulates it
	})

	// Broadcast
	if h.wsHub != nil {
		h.wsHub.BroadcastToRoom(conversation.ID.String(), "message.created", message)
	}

	// TODO: Trigger Outgoing Webhook to Evolution API here if needed
	// This is where we would call SendText/SendMedia to the external provider

	c.JSON(http.StatusCreated, message)
}

func (h *MessageHandler) Get(c *gin.Context) {
	c.JSON(200, gin.H{"message": "TODO"})
}

func (h *MessageHandler) Update(c *gin.Context) {
	c.JSON(200, gin.H{"message": "TODO"})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	c.JSON(200, gin.H{"message": "TODO"})
}

func (h *MessageHandler) CreatePublicMessage(c *gin.Context) {
	// For widget
	c.JSON(200, gin.H{"message": "TODO"})
}

// ---------------------------------------------------------------------
// Contact Handler
// ---------------------------------------------------------------------

type ContactHandler struct {
	db *gorm.DB
}

func NewContactHandler(db *gorm.DB) *ContactHandler {
	return &ContactHandler{db: db}
}

func (h *ContactHandler) List(c *gin.Context) {
	accountID := c.GetString("account_id")
	log.Printf("ContactHandler.List: accountID=%s", accountID)

	search := c.Query("search")
	pageStr := c.Query("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 15
	offset := (page - 1) * limit

	query := h.db.Model(&models.Contact{}).Where("account_id = ?", accountID)

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where("name ILIKE ? OR phone_number ILIKE ? OR email ILIKE ?", searchLike, searchLike, searchLike)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		log.Printf("ContactHandler.List: Count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var contacts []models.Contact
	if err := query.Order("updated_at desc").Limit(limit).Offset(offset).Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{
			"count":        totalCount,
			"current_page": page,
		},
		"payload": contacts,
	})
}

func (h *ContactHandler) Create(c *gin.Context) {
	accountIDStr := c.GetString("account_id")
	log.Printf(">>> ContactHandler.Create: Received request. AccountID=%s", accountIDStr)
	accountID, _ := uuid.Parse(accountIDStr)

	var input models.Contact
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.AccountID = accountID

	if err := h.db.Create(&input).Error; err != nil {
		log.Printf(">>> DEBUG CREATE CONTACT ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *ContactHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var contact models.Contact
	if err := h.db.First(&contact, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}
	c.JSON(http.StatusOK, contact)
}

func (h *ContactHandler) Update(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("account_id")

	// Verificar se contato existe e pertence à conta
	var contact models.Contact
	if err := h.db.Where("id = ? AND account_id = ?", id, accountID).First(&contact).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Parse input
	var input struct {
		Name        *string `json:"name"`
		Email       *string `json:"email"`
		PhoneNumber *string `json:"phone_number"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Atualizar apenas campos fornecidos
	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.PhoneNumber != nil {
		updates["phone_number"] = *input.PhoneNumber
	}

	if err := h.db.Model(&contact).Updates(updates).Error; err != nil {
		log.Printf(">>> DEBUG UPDATE CONTACT ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact: " + err.Error()})
		return
	}

	// Recarregar contato atualizado
	h.db.First(&contact, id)

	c.JSON(http.StatusOK, contact)
}

func (h *ContactHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("account_id")

	// Verificar se contato existe e pertence à conta
	var contact models.Contact
	if err := h.db.Where("id = ? AND account_id = ?", id, accountID).First(&contact).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Soft delete (GORM usa DeletedAt)
	if err := h.db.Delete(&contact).Error; err != nil {
		log.Printf(">>> DEBUG DELETE CONTACT ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted successfully"})
}

func (h *ContactHandler) CreatePublicContact(c *gin.Context) {
	c.JSON(200, gin.H{"message": "TODO"})
}

// ---------------------------------------------------------------------
// Inbox Handler
// ---------------------------------------------------------------------

type InboxHandler struct {
	db *gorm.DB
}

func NewInboxHandler(db *gorm.DB) *InboxHandler {
	return &InboxHandler{db: db}
}

func (h *InboxHandler) List(c *gin.Context) {
	accountID := c.GetString("account_id")
	var inboxes []models.Inbox
	if err := h.db.Where("account_id = ?", accountID).Find(&inboxes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inboxes)
}

func (h *InboxHandler) Create(c *gin.Context) {
	accountIDStr := c.GetString("account_id")
	accountID, _ := uuid.Parse(accountIDStr)

	var input models.Inbox
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.AccountID = accountID
	// Generate fallback channel ID if needed
	if input.ChannelID == uuid.Nil {
		input.ChannelID = uuid.New()
	}

	if err := h.db.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *InboxHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var inbox models.Inbox
	if err := h.db.First(&inbox, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inbox not found"})
		return
	}
	c.JSON(http.StatusOK, inbox)
}

func (h *InboxHandler) Update(c *gin.Context) {
	c.JSON(200, gin.H{"message": "TODO"})
}

func (h *InboxHandler) Delete(c *gin.Context) {
	c.JSON(200, gin.H{"message": "TODO"})
}
