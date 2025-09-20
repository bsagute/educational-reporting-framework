package handlers

import (
	"net/http"
	"time"

	"reporting-framework/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionHandler struct {
	db *gorm.DB
}

type StartSessionRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	Application string  `json:"application" binding:"required"`
	ClassroomID *string `json:"classroom_id"`
	DeviceType  string  `json:"device_type"`
	AppVersion  string  `json:"app_version"`
}

type EndSessionRequest struct {
	EndTime time.Time `json:"end_time" binding:"required"`
}

func NewSessionHandler(db *gorm.DB) *SessionHandler {
	return &SessionHandler{db: db}
}

func (h *SessionHandler) StartSession(c *gin.Context) {
	var req StartSessionRequest
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

	userID, err := uuid.Parse(req.UserID)
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

	session := models.Session{
		UserID:      userID,
		Application: req.Application,
		StartTime:   time.Now(),
		DeviceType:  req.DeviceType,
		AppVersion:  req.AppVersion,
	}

	if req.ClassroomID != nil {
		classroomID, err := uuid.Parse(*req.ClassroomID)
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
		session.ClassroomID = &classroomID
	}

	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create session",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.ID,
		"start_time": session.StartTime,
	})
}

func (h *SessionHandler) EndSession(c *gin.Context) {
	sessionID := c.Param("id")
	id, err := uuid.Parse(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid session_id format",
			},
		})
		return
	}

	var req EndSessionRequest
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

	var session models.Session
	if err := h.db.First(&session, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "Session not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve session",
				"details": err.Error(),
			},
		})
		return
	}

	// Calculate duration
	duration := int(req.EndTime.Sub(session.StartTime).Seconds())

	// Update session
	updates := map[string]interface{}{
		"end_time":         req.EndTime,
		"duration_seconds": duration,
	}

	if err := h.db.Model(&session).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update session",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Session ended successfully",
		"duration_seconds": duration,
	})
}

func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	id, err := uuid.Parse(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid session_id format",
			},
		})
		return
	}

	var session models.Session
	if err := h.db.Preload("User").Preload("Classroom").First(&session, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "Session not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve session",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, session)
}