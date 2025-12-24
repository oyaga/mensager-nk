package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nakamura/chatwoot-go/internal/config"
	"github.com/nakamura/chatwoot-go/internal/middleware"
	ws "github.com/nakamura/chatwoot-go/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Implement proper origin checking
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
	cfg *config.Config
}

func NewWebSocketHandler(hub *ws.Hub, cfg *config.Config) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, cfg: cfg}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	var userID uuid.UUID
	var err error

	// 1. Try to get token from query param (Standard Chatwoot frontend behavior)
	tokenStr := c.Query("token")
	if tokenStr != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		userID = claims.UserID
	} else {
		// 2. Fallback to user_id query param (For dev/testing tools)
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token or user_id required"})
			return
		}
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Create client
	client := &ws.Client{
		ID:     uuid.New(),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
		Rooms:  make(map[string]bool),
	}

	// Register client
	h.hub.Register <- client

	// Start pumps
	go client.WritePump()
	go client.ReadPump()
}
