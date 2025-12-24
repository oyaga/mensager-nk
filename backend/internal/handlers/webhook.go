package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nakamura/chatwoot-go/internal/models"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	db *gorm.DB
}

func NewWebhookHandler(db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{db: db}
}

// List all webhooks for the account
func (h *WebhookHandler) List(c *gin.Context) {
	accountID := c.GetString("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID required"})
		return
	}

	var webhooks []models.Webhook
	if err := h.db.Where("account_id = ?", accountID).Find(&webhooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// Create a new webhook
func (h *WebhookHandler) Create(c *gin.Context) {
	accountID := c.GetString("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID required"})
		return
	}

	var input struct {
		Name          string   `json:"name"`
		URL           string   `json:"url" binding:"required"`
		InboxID       string   `json:"inbox_id"`
		Subscriptions []string `json:"subscriptions"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accUUID, _ := uuid.Parse(accountID)

	webhook := models.Webhook{
		AccountID:     accUUID,
		Name:          input.Name,
		URL:           input.URL,
		Subscriptions: input.Subscriptions,
		WebhookType:   "account",
	}

	if input.InboxID != "" {
		inboxUUID, _ := uuid.Parse(input.InboxID)
		webhook.InboxID = &inboxUUID
		webhook.WebhookType = "inbox"
	}

	if err := h.db.Create(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// Update a webhook
func (h *WebhookHandler) Update(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("account_id")

	var webhook models.Webhook
	if err := h.db.Where("id = ? AND account_id = ?", id, accountID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	var input struct {
		Name          string   `json:"name"`
		URL           string   `json:"url"`
		Subscriptions []string `json:"subscriptions"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		webhook.Name = input.Name
	}
	if input.URL != "" {
		webhook.URL = input.URL
	}
	if input.Subscriptions != nil {
		webhook.Subscriptions = input.Subscriptions
	}

	if err := h.db.Save(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// Delete a webhook
func (h *WebhookHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("account_id")

	result := h.db.Where("id = ? AND account_id = ?", id, accountID).Delete(&models.Webhook{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted"})
}

// DispatchEvent sends event to all matching webhooks
func (h *WebhookHandler) DispatchEvent(accountID uuid.UUID, eventName string, payload interface{}) {
	var webhooks []models.Webhook
	h.db.Where("account_id = ?", accountID).Find(&webhooks)

	for _, webhook := range webhooks {
		// Check if webhook subscribes to this event
		if !containsSubscription(webhook.Subscriptions, eventName) && len(webhook.Subscriptions) > 0 {
			continue
		}

		go sendWebhookRequest(webhook.URL, eventName, payload)
	}
}

func containsSubscription(subscriptions []string, event string) bool {
	for _, s := range subscriptions {
		if s == event || s == "*" {
			return true
		}
	}
	return false
}

func sendWebhookRequest(url string, eventName string, payload interface{}) {
	body := map[string]interface{}{
		"event":     eventName,
		"timestamp": time.Now().Unix(),
		"data":      payload,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nakamura-Event", eventName)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
