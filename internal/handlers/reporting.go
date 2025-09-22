package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"reporting-framework/internal/domain/reporting"
)

// ReportingHandler handles reporting-related HTTP requests
type ReportingHandler struct {
	db *gorm.DB
}

// NewReportingHandler creates a new reporting handler
func NewReportingHandler(db *gorm.DB) *ReportingHandler {
	return &ReportingHandler{
		db: db,
	}
}

// RegisterRoutes registers all reporting routes
func (h *ReportingHandler) RegisterRoutes(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		// Event ingestion endpoints
		v1.POST("/events", h.IngestEvents)
		v1.POST("/sessions/batch", h.IngestSessionBatch)

		// Report generation endpoints
		reports := v1.Group("/reports")
		{
			reports.GET("/student-performance", h.GetStudentPerformanceReport)
			reports.GET("/classroom-engagement", h.GetClassroomEngagementReport)
			reports.GET("/content-effectiveness", h.GetContentEffectivenessReport)
			reports.GET("/school-overview", h.GetSchoolOverviewReport)
		}

		// Analytics endpoints
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/real-time/active-sessions", h.GetActiveSessions)
			analytics.GET("/trends/engagement", h.GetEngagementTrends)
			analytics.GET("/quiz-analytics/:quiz_id", h.GetQuizAnalytics)
		}

		// Generic query endpoint (cube.dev style)
		v1.POST("/query", h.ExecuteGenericQuery)

		// Administrative endpoints
		admin := v1.Group("/admin")
		{
			admin.POST("/schools", h.CreateSchool)
			admin.POST("/classrooms", h.CreateClassroom)
			admin.POST("/users", h.CreateUser)
			admin.POST("/refresh-metrics", h.RefreshAggregatedMetrics)
		}
	}
}

// IngestEvents handles batch event ingestion
func (h *ReportingHandler) IngestEvents(c *gin.Context) {
	var req reporting.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Get user context from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	schoolID, _ := c.Get("school_id")

	var events []reporting.Event
	var eventIDs []uuid.UUID

	for _, eventData := range req.Events {
		event := reporting.Event{
			ID:          uuid.New(),
			EventType:   eventData.EventType,
			UserID:      eventData.UserID,
			SessionID:   eventData.SessionID,
			ClassroomID: eventData.ClassroomID,
			Application: eventData.Application,
			Timestamp:   eventData.Timestamp,
			CreatedAt:   time.Now(),
		}

		// Set user and school context if not provided
		if event.UserID == nil {
			if uid, ok := userID.(uuid.UUID); ok {
				event.UserID = &uid
			}
		}
		if event.SchoolID == nil {
			if sid, ok := schoolID.(uuid.UUID); ok {
				event.SchoolID = &sid
			}
		}

		// Convert metadata and device info
		if eventData.Metadata != nil {
			event.Metadata = reporting.JSONB(eventData.Metadata)
		}
		if eventData.DeviceInfo != nil {
			event.DeviceInfo = reporting.JSONB(eventData.DeviceInfo)
		}

		events = append(events, event)
		eventIDs = append(eventIDs, event.ID)
	}

	// Batch insert events
	if err := h.db.CreateInBatches(events, 100).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store events", "details": err.Error()})
		return
	}

	// Trigger async aggregation update (in a real system, this would be done via message queue)
	go h.updateAggregatedMetrics(events)

	response := reporting.EventResponse{
		Success:        true,
		ProcessedCount: len(events),
		Message:        "Events ingested successfully",
		EventIDs:       eventIDs,
	}

	c.JSON(http.StatusCreated, response)
}

// IngestSessionBatch handles batch session data ingestion
func (h *ReportingHandler) IngestSessionBatch(c *gin.Context) {
	var req struct {
		Sessions []struct {
			Application string               `json:"application"`
			StartTime   time.Time            `json:"start_time"`
			EndTime     *time.Time           `json:"end_time"`
			ClassroomID *uuid.UUID           `json:"classroom_id"`
			DeviceInfo  map[string]interface{} `json:"device_info"`
			Events      []reporting.EventData `json:"events"`
		} `json:"sessions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid := userID.(uuid.UUID)
	var processedSessions []uuid.UUID

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, sessionData := range req.Sessions {
		session := reporting.Session{
			ID:          uuid.New(),
			UserID:      uid,
			ClassroomID: sessionData.ClassroomID,
			Application: sessionData.Application,
			StartTime:   sessionData.StartTime,
			EndTime:     sessionData.EndTime,
			CreatedAt:   time.Now(),
		}

		if sessionData.EndTime != nil {
			duration := int(sessionData.EndTime.Sub(sessionData.StartTime).Seconds())
			session.DurationSeconds = &duration
		}

		if sessionData.DeviceInfo != nil {
			session.DeviceInfo = reporting.JSONB(sessionData.DeviceInfo)
		}

		if err := tx.Create(&session).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session", "details": err.Error()})
			return
		}

		// Create associated events
		for _, eventData := range sessionData.Events {
			event := reporting.Event{
				ID:          uuid.New(),
				EventType:   eventData.EventType,
				UserID:      &uid,
				SessionID:   &session.ID,
				ClassroomID: sessionData.ClassroomID,
				Application: &sessionData.Application,
				Timestamp:   eventData.Timestamp,
				CreatedAt:   time.Now(),
			}

			if eventData.Metadata != nil {
				event.Metadata = reporting.JSONB(eventData.Metadata)
			}

			if err := tx.Create(&event).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event", "details": err.Error()})
				return
			}
		}

		processedSessions = append(processedSessions, session.ID)
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":           true,
		"processed_sessions": len(processedSessions),
		"session_ids":       processedSessions,
		"message":           "Sessions ingested successfully",
	})
}

// Date format constants
const (
	DateFormat = "2006-01-02"
	OrderByDateASC = "date ASC"
)

// GetStudentPerformanceReport generates student performance analytics
func (h *ReportingHandler) GetStudentPerformanceReport(c *gin.Context) {
	studentIDStr := c.Query("student_id")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	includeDetails := c.Query("include_details") == "true"

	if studentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student_id is required"})
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student_id format"})
		return
	}

	dateFrom, dateTo, err := h.parseDateRange(dateFromStr, dateToStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build base query
	query := h.db.Table("daily_user_metrics dum").
		Select(`
			dum.user_id,
			AVG(dum.avg_quiz_score) as avg_quiz_score,
			SUM(dum.quiz_attempts) as total_quiz_attempts,
			SUM(dum.quiz_completions) as total_quiz_completions,
			AVG(dum.total_session_duration_seconds / 60.0) as avg_daily_minutes,
			SUM(dum.events_count) as total_events,
			COUNT(dum.date) as active_days
		`).
		Where("dum.user_id = ? AND dum.date BETWEEN ? AND ?", studentID, dateFrom, dateTo).
		Group("dum.user_id")

	var result struct {
		UserID               uuid.UUID `json:"user_id"`
		AvgQuizScore         *float64  `json:"avg_quiz_score"`
		TotalQuizAttempts    int       `json:"total_quiz_attempts"`
		TotalQuizCompletions int       `json:"total_quiz_completions"`
		AvgDailyMinutes      float64   `json:"avg_daily_minutes"`
		TotalEvents          int       `json:"total_events"`
		ActiveDays           int       `json:"active_days"`
	}

	if err := query.Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance data", "details": err.Error()})
		return
	}

	// Calculate engagement score
	engagementScore := calculateEngagementScore(result.AvgDailyMinutes, result.ActiveDays, float64(dateTo.Sub(dateFrom).Hours()/24))

	response := gin.H{
		"student_id": studentID,
		"period":     gin.H{"from": dateFrom.Format(DateFormat), "to": dateTo.Format(DateFormat)},
		"overall_stats": gin.H{
			"avg_quiz_score":       result.AvgQuizScore,
			"total_quiz_attempts":  result.TotalQuizAttempts,
			"total_quiz_completions": result.TotalQuizCompletions,
			"avg_daily_minutes":    result.AvgDailyMinutes,
			"total_events":         result.TotalEvents,
			"active_days":          result.ActiveDays,
			"engagement_score":     engagementScore,
		},
	}

	if includeDetails {
		// Add detailed quiz performance
		var quizPerformance []gin.H
		h.db.Table("quiz_sessions qs").
			Select("q.title, qs.percentage_score, qs.completed_at, qs.time_spent_seconds").
			Joins("JOIN quizzes q ON qs.quiz_id = q.id").
			Where("qs.student_id = ? AND qs.completed_at BETWEEN ? AND ? AND qs.is_completed = true",
				studentID, dateFrom, dateTo).
			Order("qs.completed_at DESC").
			Scan(&quizPerformance)

		response["quiz_performance"] = quizPerformance

		// Add learning progression (daily metrics over time)
		var learningProgression []gin.H
		h.db.Table("daily_user_metrics").
			Select("date, avg_quiz_score, total_session_duration_seconds / 60 as daily_minutes, events_count").
			Where("user_id = ? AND date BETWEEN ? AND ?", studentID, dateFrom, dateTo).
			Order(OrderByDateASC).
			Scan(&learningProgression)

		response["learning_progression"] = learningProgression
	}

	c.JSON(http.StatusOK, response)
}

// GetClassroomEngagementReport generates classroom engagement analytics
func (h *ReportingHandler) GetClassroomEngagementReport(c *gin.Context) {
	classroomIDStr := c.Query("classroom_id")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	if classroomIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "classroom_id is required"})
		return
	}

	classroomID, err := uuid.Parse(classroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom_id format"})
		return
	}

	// Parse date range (default to last 30 days)
	dateFrom, dateTo, err := h.parseDateRangeWithDefault(dateFromStr, dateToStr, -30)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get classroom engagement metrics
	var engagementMetrics struct {
		ActiveParticipationRate  *float64 `json:"active_participation_rate"`
		AvgSessionDuration       float64  `json:"avg_session_duration"`
		CollaborationEvents      int      `json:"collaboration_events"`
		ContentSharingFrequency  float64  `json:"content_sharing_frequency"`
		TotalQuizSessions        int      `json:"total_quiz_sessions"`
		AvgQuizCompletionRate    *float64 `json:"avg_quiz_completion_rate"`
	}

	h.db.Table("daily_classroom_metrics").
		Select(`
			AVG(participation_rate) as active_participation_rate,
			AVG(avg_session_duration_minutes) as avg_session_duration,
			SUM(sync_events_count) as collaboration_events,
			AVG(content_shared_count) as content_sharing_frequency,
			SUM(total_quiz_sessions) as total_quiz_sessions,
			AVG(avg_quiz_completion_rate) as avg_quiz_completion_rate
		`).
		Where("classroom_id = ? AND date BETWEEN ? AND ?", classroomID, dateFrom, dateTo).
		Scan(&engagementMetrics)

	// Get student breakdown
	var studentBreakdown []gin.H
	h.db.Table("users u").
		Select(`
			u.id, u.first_name, u.last_name,
			AVG(dum.avg_quiz_score) as avg_quiz_score,
			AVG(dum.total_session_duration_seconds / 60.0) as avg_daily_minutes,
			COUNT(dum.date) as active_days
		`).
		Joins("JOIN user_classrooms uc ON u.id = uc.user_id").
		Joins("LEFT JOIN daily_user_metrics dum ON u.id = dum.user_id AND dum.date BETWEEN ? AND ?", dateFrom, dateTo).
		Where("uc.classroom_id = ? AND uc.is_active = true AND u.role = 'student'", classroomID).
		Group("u.id, u.first_name, u.last_name").
		Scan(&studentBreakdown)

	// Get timeline data (daily engagement over time)
	var timelineData []gin.H
	h.db.Table("daily_classroom_metrics").
		Select("date, active_students_count, participation_rate, avg_session_duration_minutes, engagement_score").
		Where("classroom_id = ? AND date BETWEEN ? AND ?", classroomID, dateFrom, dateTo).
		Order(OrderByDateASC).
		Scan(&timelineData)

	response := gin.H{
		"classroom_id":        classroomID,
		"period":             gin.H{"from": dateFrom.Format(DateFormat), "to": dateTo.Format(DateFormat)},
		"engagement_metrics":  engagementMetrics,
		"student_breakdown":   studentBreakdown,
		"timeline_data":       timelineData,
	}

	c.JSON(http.StatusOK, response)
}

// GetContentEffectivenessReport generates content effectiveness analytics
func (h *ReportingHandler) GetContentEffectivenessReport(c *gin.Context) {
	schoolIDStr := c.Query("school_id")
	classroomIDStr := c.Query("classroom_id")
	contentType := c.Query("content_type")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	// Parse filters
	var schoolID, classroomID *uuid.UUID
	if schoolIDStr != "" {
		if id, err := uuid.Parse(schoolIDStr); err == nil {
			schoolID = &id
		}
	}
	if classroomIDStr != "" {
		if id, err := uuid.Parse(classroomIDStr); err == nil {
			classroomID = &id
		}
	}

	var dateFrom, dateTo time.Time
	var err error
	dateFrom, dateTo, err = h.parseDateRangeWithDefault(dateFromStr, dateToStr, -30)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build query for content analytics
	query := h.db.Table("content c").
		Select(`
			c.content_type,
			COUNT(c.id) as total_content,
			AVG(cm.view_count) as avg_views,
			AVG(cm.unique_viewers) as avg_unique_viewers,
			AVG(cm.avg_view_duration_seconds) as avg_view_duration,
			AVG(cm.effectiveness_score) as avg_effectiveness_score
		`).
		Joins("LEFT JOIN content_metrics cm ON c.id = cm.content_id").
		Where("c.created_at BETWEEN ? AND ?", dateFrom, dateTo)

	if schoolID != nil {
		query = query.Joins("JOIN classrooms cl ON c.classroom_id = cl.id").
			Where("cl.school_id = ?", *schoolID)
	}
	if classroomID != nil {
		query = query.Where("c.classroom_id = ?", *classroomID)
	}
	if contentType != "" {
		query = query.Where("c.content_type = ?", contentType)
	}

	var contentAnalytics []gin.H
	query.Group("c.content_type").Scan(&contentAnalytics)

	// Get most engaging content
	mostEngagingQuery := h.db.Table("content c").
		Select("c.title, c.content_type, cm.view_count, cm.effectiveness_score, c.created_at").
		Joins("JOIN content_metrics cm ON c.id = cm.content_id").
		Where("c.created_at BETWEEN ? AND ?", dateFrom, dateTo)

	if classroomID != nil {
		mostEngagingQuery = mostEngagingQuery.Where("c.classroom_id = ?", *classroomID)
	}

	var mostEngagingContent []gin.H
	mostEngagingQuery.Order("cm.effectiveness_score DESC").Limit(10).Scan(&mostEngagingContent)

	response := gin.H{
		"period": gin.H{"from": dateFrom.Format(DateFormat), "to": dateTo.Format(DateFormat)},
		"content_analytics": gin.H{
			"content_type_breakdown":   contentAnalytics,
			"most_engaging_content":    mostEngagingContent,
		},
		"recommendations": []string{
			"Focus on creating more interactive content types with higher engagement scores",
			"Review content with low view duration for potential improvements",
			"Encourage content sharing to increase reach and effectiveness",
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetSchoolOverviewReport generates high-level school analytics
func (h *ReportingHandler) GetSchoolOverviewReport(c *gin.Context) {
	schoolIDStr := c.Query("school_id")
	if schoolIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school_id is required"})
		return
	}

	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid school_id format"})
		return
	}

	// Get latest weekly metrics
	var weeklyMetrics struct {
		TotalClassrooms     int     `json:"total_classrooms"`
		ActiveClassrooms    int     `json:"active_classrooms"`
		TotalStudents       int     `json:"total_students"`
		ActiveStudents      int     `json:"active_students"`
		TotalTeachers       int     `json:"total_teachers"`
		ActiveTeachers      int     `json:"active_teachers"`
		AvgSchoolEngagement float64 `json:"avg_school_engagement"`
		PlatformAdoptionRate float64 `json:"platform_adoption_rate"`
	}

	h.db.Table("weekly_school_metrics").
		Where("school_id = ?", schoolID).
		Order("week_start_date DESC").
		Limit(1).
		Scan(&weeklyMetrics)

	c.JSON(http.StatusOK, gin.H{
		"school_id": schoolID,
		"overview":  weeklyMetrics,
		"timestamp": time.Now(),
	})
}

// ExecuteGenericQuery handles cube.dev style queries
func (h *ReportingHandler) ExecuteGenericQuery(c *gin.Context) {
	var queryReq struct {
		Measures       []string `json:"measures"`
		Dimensions     []string `json:"dimensions"`
		TimeDimensions []struct {
			Dimension string   `json:"dimension"`
			Granularity string `json:"granularity"`
			DateRange   []string `json:"dateRange"`
		} `json:"timeDimensions"`
		Filters []struct {
			Member   string   `json:"member"`
			Operator string   `json:"operator"`
			Values   []string `json:"values"`
		} `json:"filters"`
		Order [][]string `json:"order"`
		Limit int        `json:"limit"`
	}

	if err := c.ShouldBindJSON(&queryReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query format", "details": err.Error()})
		return
	}

	// This is a simplified implementation - in a real system, you'd have a proper query builder
	// For demo purposes, return a sample result
	result := []map[string]interface{}{
		{
			"users_role":        "student",
			"events_event_type": "quiz_answer_submitted",
			"events_count":      125,
		},
		{
			"users_role":        "teacher",
			"events_event_type": "quiz_created",
			"events_count":      15,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"query": queryReq,
		"executedAt": time.Now(),
	})
}

// Helper functions

func calculateEngagementScore(avgDailyMinutes float64, activeDays int, totalDays float64) float64 {
	if totalDays == 0 {
		return 0
	}

	// Simple engagement scoring algorithm
	consistencyScore := float64(activeDays) / totalDays * 100
	intensityScore := avgDailyMinutes / 60 * 100 // Normalize to hours

	// Weight consistency more heavily than intensity
	return (consistencyScore * 0.7) + (intensityScore * 0.3)
}

func (h *ReportingHandler) updateAggregatedMetrics(events []reporting.Event) {
	// This would typically be handled by a background job or message queue
	// For demo purposes, we'll do basic aggregation updates

	// Group events by user and date
	userDates := make(map[string]map[string]int)

	for _, event := range events {
		if event.UserID == nil {
			continue
		}

		userKey := event.UserID.String()
		dateKey := event.Timestamp.Format(DateFormat)

		if userDates[userKey] == nil {
			userDates[userKey] = make(map[string]int)
		}
		userDates[userKey][dateKey]++
	}

	// Update daily user metrics
	for userKey, dates := range userDates {
		userID, _ := uuid.Parse(userKey)
		for dateKey, eventCount := range dates {
			date, _ := time.Parse(DateFormat, dateKey)

			// Upsert daily metrics
			h.db.Exec(`
				INSERT INTO daily_user_metrics (user_id, date, events_count, created_at, updated_at)
				VALUES (?, ?, ?, NOW(), NOW())
				ON CONFLICT (user_id, date)
				DO UPDATE SET
					events_count = daily_user_metrics.events_count + ?,
					updated_at = NOW()
			`, userID, date, eventCount, eventCount)
		}
	}
}

// Additional helper functions for different report types...

// CreateSchool - Admin endpoint to create schools
func (h *ReportingHandler) CreateSchool(c *gin.Context) {
	var school reporting.School
	if err := c.ShouldBindJSON(&school); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&school).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create school"})
		return
	}

	c.JSON(http.StatusCreated, school)
}

// CreateClassroom - Admin endpoint to create classrooms
func (h *ReportingHandler) CreateClassroom(c *gin.Context) {
	var classroom reporting.Classroom
	if err := c.ShouldBindJSON(&classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&classroom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create classroom"})
		return
	}

	c.JSON(http.StatusCreated, classroom)
}

// CreateUser - Admin endpoint to create users
func (h *ReportingHandler) CreateUser(c *gin.Context) {
	var user reporting.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// RefreshAggregatedMetrics - Admin endpoint to refresh aggregated metrics
func (h *ReportingHandler) RefreshAggregatedMetrics(c *gin.Context) {
	// Refresh materialized view
	if err := h.db.Exec("SELECT refresh_classroom_performance_mv()").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Metrics refreshed successfully"})
}

// GetActiveSessions - Real-time active sessions
func (h *ReportingHandler) GetActiveSessions(c *gin.Context) {
	schoolIDStr := c.Query("school_id")
	var sessions []gin.H

	query := h.db.Table("active_sessions ass").
		Select("ass.*, u.first_name, u.last_name, c.name as classroom_name").
		Joins("JOIN users u ON ass.user_id = u.id").
		Joins("LEFT JOIN classrooms c ON ass.classroom_id = c.id").
		Where("ass.last_heartbeat > ?", time.Now().Add(-5*time.Minute))

	if schoolIDStr != "" {
		query = query.Where("u.school_id = ?", schoolIDStr)
	}

	query.Scan(&sessions)
	c.JSON(http.StatusOK, gin.H{"active_sessions": sessions})
}

// GetEngagementTrends - Engagement trends over time
func (h *ReportingHandler) GetEngagementTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "7d") // 7d, 30d, 90d
	schoolIDStr := c.Query("school_id")

	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	dateFrom := time.Now().AddDate(0, 0, -days)

	query := h.db.Table("daily_classroom_metrics").
		Select("date, AVG(engagement_score) as avg_engagement, SUM(active_students_count) as total_active_students").
		Where("date >= ?", dateFrom).
		Group("date").
		Order(OrderByDateASC)

	if schoolIDStr != "" {
		query = query.Joins("JOIN classrooms c ON daily_classroom_metrics.classroom_id = c.id").
			Where("c.school_id = ?", schoolIDStr)
	}

	var trends []gin.H
	query.Scan(&trends)

	c.JSON(http.StatusOK, gin.H{
		"period": period,
		"trends": trends,
	})
}

// GetQuizAnalytics - Detailed quiz analytics
func (h *ReportingHandler) GetQuizAnalytics(c *gin.Context) {
	quizIDStr := c.Param("quiz_id")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz_id"})
		return
	}

	var analytics gin.H
	h.db.Table("quiz_analytics").Where("quiz_id = ?", quizID).Scan(&analytics)

	if len(analytics) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// parseDateRange parses date range from query parameters
func (h *ReportingHandler) parseDateRange(dateFromStr, dateToStr string) (time.Time, time.Time, error) {
	var dateFrom, dateTo time.Time
	var err error

	if dateFromStr != "" {
		dateFrom, err = time.Parse(DateFormat, dateFromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date_from format (YYYY-MM-DD)")
		}
	} else {
		dateFrom = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if dateToStr != "" {
		dateTo, err = time.Parse(DateFormat, dateToStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date_to format (YYYY-MM-DD)")
		}
	} else {
		dateTo = time.Now()
	}

	return dateFrom, dateTo, nil
}

// parseDateRangeWithDefault parses date range with a default day offset
func (h *ReportingHandler) parseDateRangeWithDefault(dateFromStr, dateToStr string, defaultDays int) (time.Time, time.Time, error) {
	var dateFrom, dateTo time.Time
	var err error

	if dateFromStr != "" {
		dateFrom, err = time.Parse(DateFormat, dateFromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date_from format (YYYY-MM-DD)")
		}
	} else {
		dateFrom = time.Now().AddDate(0, 0, defaultDays)
	}

	if dateToStr != "" {
		dateTo, err = time.Parse(DateFormat, dateToStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date_to format (YYYY-MM-DD)")
		}
	} else {
		dateTo = time.Now()
	}

	return dateFrom, dateTo, nil
}