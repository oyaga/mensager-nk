package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nakamura/chatwoot-go/internal/models"
	"gorm.io/gorm"
)

type AccountHandler struct {
	db *gorm.DB
}

func NewAccountHandler(db *gorm.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

func (h *AccountHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "List accounts - TODO"})
}

func (h *AccountHandler) Create(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create account - TODO"})
}

func (h *AccountHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get account - TODO"})
}

func (h *AccountHandler) Update(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update account - TODO"})
}

func (h *AccountHandler) Delete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete account - TODO"})
}

func (h *AccountHandler) ListUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "List account users - TODO"})
}

func (h *AccountHandler) AddUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Add user to account - TODO"})
}

func (h *AccountHandler) RemoveUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Remove user from account - TODO"})
}

func (h *AccountHandler) GetStats(c *gin.Context) {
	accountID := c.MustGet("account_id").(uuid.UUID)

	var stats struct {
		TotalConversations int64 `json:"total_conversations"`
		OpenConversations  int64 `json:"open_conversations"`
		TotalContacts      int64 `json:"total_contacts"`
		TotalMessages      int64 `json:"total_messages"`
	}

	h.db.Model(&models.Conversation{}).Where("account_id = ?", accountID).Count(&stats.TotalConversations)
	h.db.Model(&models.Conversation{}).Where("account_id = ? AND status = ?", accountID, "open").Count(&stats.OpenConversations)
	h.db.Model(&models.Contact{}).Where("account_id = ?", accountID).Count(&stats.TotalContacts)
	h.db.Model(&models.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Where("conversations.account_id = ?", accountID).
		Count(&stats.TotalMessages)

	c.JSON(http.StatusOK, stats)
}
