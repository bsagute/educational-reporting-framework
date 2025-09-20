package handlers

import (
	"net/http"
	"time"

	"reporting-framework/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventHandler struct {
	db *gorm.DB
}

type EventRequest struct {
	Events []EventData `json:"events" binding:"required"`
}

type EventData struct {
	EventType   string                 `json:"event_type" binding:"required"`
	Timestamp   time.Time              `json:"timestamp" binding:"required"`
	UserID      string                 `json:"user_id" binding:"required"`
	SessionID   string                 `json:"session_id" binding:"required"`
	ClassroomID *string                `json:"classroom_id"`
	Application string                 `json:"application" binding:"required"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{db: db}
}

func (h *EventHandler) BatchInsert(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	// Convert request to models
	events := make([]models.Event, len(req.Events))
	for i, eventData := range req.Events {
		userID, err := uuid.Parse(eventData.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": map[string]interface{}{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid user_id format",
					"details": "user_id must be a valid UUID",
				},
			})
			return
		}

		sessionID, err := uuid.Parse(eventData.SessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": map[string]interface{}{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid session_id format",
					"details": "session_id must be a valid UUID",
				},
			})
			return
		}

		event := models.Event{
			EventType:   eventData.EventType,
			UserID:      userID,
			SessionID:   sessionID,
			Timestamp:   eventData.Timestamp,
			Application: eventData.Application,
			Payload:     models.JSONB(eventData.Payload),
			Metadata:    models.JSONB(eventData.Metadata),
		}

		if eventData.ClassroomID != nil {
			classroomID, err := uuid.Parse(*eventData.ClassroomID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": map[string]interface{}{
						"code":    "VALIDATION_ERROR",
						"message": "Invalid classroom_id format",
						"details": "classroom_id must be a valid UUID",
					},
				})
				return
			}
			event.ClassroomID = &classroomID
		}

		events[i] = event
	}

	// Batch insert events
	if err := h.db.CreateInBatches(events, 100).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to insert events",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Events inserted successfully",
		"events_created": len(events),
	})
}