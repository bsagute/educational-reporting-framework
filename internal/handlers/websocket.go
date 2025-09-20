package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type ClassroomLiveData struct {
	ActiveStudents        int                    `json:"active_students"`
	QuizParticipationRate float64                `json:"quiz_participation_rate"`
	AverageResponseTime   float64                `json:"average_response_time"`
	EngagementScore       float64                `json:"engagement_score"`
	RecentEvents          []map[string]interface{} `json:"recent_events"`
}

func HandleWebSocket(c *gin.Context, db *gorm.DB) {
	classroomID := c.Param("id")
	id, err := uuid.Parse(classroomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid classroom_id format",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Send initial data
	initialData, err := getClassroomLiveData(db, id)
	if err != nil {
		log.Printf("Failed to get initial data: %v", err)
		return
	}

	initialMessage := WebSocketMessage{
		Type:      "initial_data",
		Data:      map[string]interface{}{"classroom_data": initialData},
		Timestamp: time.Now(),
	}

	if err := conn.WriteJSON(initialMessage); err != nil {
		log.Printf("Failed to send initial message: %v", err)
		return
	}

	// Start periodic updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Channel to handle connection close
	done := make(chan struct{})

	// Read messages from client (for keepalive)
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}
		}
	}()

	// Send periodic updates
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Get updated classroom data
			liveData, err := getClassroomLiveData(db, id)
			if err != nil {
				log.Printf("Failed to get live data: %v", err)
				continue
			}

			message := WebSocketMessage{
				Type:      "live_update",
				Data:      map[string]interface{}{"classroom_data": liveData},
				Timestamp: time.Now(),
			}

			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to send update: %v", err)
				return
			}
		}
	}
}

func getClassroomLiveData(db *gorm.DB, classroomID uuid.UUID) (*ClassroomLiveData, error) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	// Get active students in the last hour
	var activeStudents int64
	err := db.Table("sessions").
		Distinct("user_id").
		Where("classroom_id = ?", classroomID).
		Where("start_time >= ?", oneHourAgo).
		Count(&activeStudents)
	if err != nil {
		return nil, err
	}

	// Get quiz participation data
	var quizData struct {
		TotalQuizzes       int `json:"total_quizzes"`
		ParticipatingUsers int `json:"participating_users"`
		TotalResponses     int `json:"total_responses"`
		AvgResponseTime    float64 `json:"avg_response_time"`
	}

	err = db.Table("quizzes").
		Select(`
			COUNT(DISTINCT quizzes.id) as total_quizzes,
			COUNT(DISTINCT quiz_responses.student_id) as participating_users,
			COUNT(quiz_responses.id) as total_responses,
			AVG(quiz_responses.time_taken_seconds) as avg_response_time
		`).
		Joins("LEFT JOIN quiz_responses ON quizzes.id = quiz_responses.quiz_id").
		Where("quizzes.classroom_id = ?", classroomID).
		Where("quizzes.created_at >= ?", oneHourAgo).
		Scan(&quizData).Error
	if err != nil {
		return nil, err
	}

	// Calculate participation rate
	participationRate := 0.0
	if activeStudents > 0 {
		participationRate = float64(quizData.ParticipatingUsers) / float64(activeStudents) * 100
	}

	// Calculate engagement score
	engagementScore := (participationRate + float64(quizData.TotalResponses)*2) / 3

	// Get recent events
	var recentEvents []map[string]interface{}
	err = db.Table("events").
		Select("event_type, timestamp, payload").
		Where("classroom_id = ?", classroomID).
		Where("timestamp >= ?", now.Add(-10*time.Minute)).
		Order("timestamp DESC").
		Limit(10).
		Scan(&recentEvents).Error
	if err != nil {
		return nil, err
	}

	return &ClassroomLiveData{
		ActiveStudents:        int(activeStudents),
		QuizParticipationRate: participationRate,
		AverageResponseTime:   quizData.AvgResponseTime,
		EngagementScore:       engagementScore,
		RecentEvents:          recentEvents,
	}, nil
}