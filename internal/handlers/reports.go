package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"reporting-framework/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db *gorm.DB
}

type StudentPerformanceReport struct {
	StudentID string     `json:"student_id"`
	Period    DatePeriod `json:"period"`
	Metrics   StudentMetrics `json:"metrics"`
}

type DatePeriod struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type StudentMetrics struct {
	QuizPerformance QuizPerformanceMetrics `json:"quiz_performance"`
	Engagement      EngagementMetrics      `json:"engagement"`
}

type QuizPerformanceMetrics struct {
	TotalQuizzes      int                    `json:"total_quizzes"`
	AverageScore      float64                `json:"average_score"`
	ImprovementTrend  float64                `json:"improvement_trend"`
	SubjectBreakdown  []SubjectPerformance   `json:"subject_breakdown"`
}

type SubjectPerformance struct {
	Subject      string  `json:"subject"`
	AverageScore float64 `json:"average_score"`
	QuizCount    int     `json:"quiz_count"`
}

type EngagementMetrics struct {
	SessionCount           int     `json:"session_count"`
	TotalTimeMinutes       float64 `json:"total_time_minutes"`
	AverageSessionDuration float64 `json:"average_session_duration"`
}

type ClassroomEngagementReport struct {
	ClassroomID string                    `json:"classroom_id"`
	Date        string                    `json:"date"`
	Metrics     ClassroomEngagementMetrics `json:"metrics"`
}

type ClassroomEngagementMetrics struct {
	ActiveStudents         int     `json:"active_students"`
	QuizParticipationRate  float64 `json:"quiz_participation_rate"`
	AverageResponseTime    float64 `json:"average_response_time"`
	EngagementScore        float64 `json:"engagement_score"`
	TotalInteractions      int     `json:"total_interactions"`
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

func (h *ReportHandler) GetStudentPerformance(c *gin.Context) {
	studentID := c.Param("id")
	id, err := uuid.Parse(studentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid student_id format",
			},
		})
		return
	}

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	subject := c.Query("subject")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "start_date and end_date are required",
			},
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid start_date format. Use YYYY-MM-DD",
			},
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid end_date format. Use YYYY-MM-DD",
			},
		})
		return
	}

	// Get quiz performance data
	quizPerformance, err := h.getQuizPerformance(id, start, end, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to calculate quiz performance",
				"details": err.Error(),
			},
		})
		return
	}

	// Get engagement data
	engagement, err := h.getEngagementMetrics(id, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to calculate engagement metrics",
				"details": err.Error(),
			},
		})
		return
	}

	report := StudentPerformanceReport{
		StudentID: studentID,
		Period: DatePeriod{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Metrics: StudentMetrics{
			QuizPerformance: *quizPerformance,
			Engagement:      *engagement,
		},
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) getQuizPerformance(studentID uuid.UUID, start, end time.Time, subject string) (*QuizPerformanceMetrics, error) {
	// Base query for quiz responses
	query := h.db.Table("quiz_responses").
		Select(`
			COUNT(DISTINCT quiz_responses.quiz_id) as total_quizzes,
			AVG(CASE WHEN quiz_responses.points_earned IS NOT NULL THEN
				(quiz_responses.points_earned / quiz_questions.points) * 100
			END) as average_score
		`).
		Joins("JOIN quiz_questions ON quiz_responses.question_id = quiz_questions.id").
		Joins("JOIN quizzes ON quiz_responses.quiz_id = quizzes.id").
		Where("quiz_responses.student_id = ?", studentID).
		Where("quiz_responses.submitted_at BETWEEN ? AND ?", start, end)

	if subject != "" {
		query = query.Joins("JOIN classrooms ON quizzes.classroom_id = classrooms.id").
			Where("classrooms.subject = ?", subject)
	}

	var result struct {
		TotalQuizzes int     `json:"total_quizzes"`
		AverageScore float64 `json:"average_score"`
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}

	// Get subject breakdown
	subjectQuery := h.db.Table("quiz_responses").
		Select(`
			classrooms.subject,
			COUNT(DISTINCT quiz_responses.quiz_id) as quiz_count,
			AVG(CASE WHEN quiz_responses.points_earned IS NOT NULL THEN
				(quiz_responses.points_earned / quiz_questions.points) * 100
			END) as average_score
		`).
		Joins("JOIN quiz_questions ON quiz_responses.question_id = quiz_questions.id").
		Joins("JOIN quizzes ON quiz_responses.quiz_id = quizzes.id").
		Joins("JOIN classrooms ON quizzes.classroom_id = classrooms.id").
		Where("quiz_responses.student_id = ?", studentID).
		Where("quiz_responses.submitted_at BETWEEN ? AND ?", start, end).
		Group("classrooms.subject")

	var subjectBreakdown []SubjectPerformance
	if err := subjectQuery.Scan(&subjectBreakdown).Error; err != nil {
		return nil, err
	}

	// Calculate improvement trend (simplified - comparing first and last week)
	improvementTrend := 0.0 // TODO: Implement trend calculation

	return &QuizPerformanceMetrics{
		TotalQuizzes:     result.TotalQuizzes,
		AverageScore:     result.AverageScore,
		ImprovementTrend: improvementTrend,
		SubjectBreakdown: subjectBreakdown,
	}, nil
}

func (h *ReportHandler) getEngagementMetrics(studentID uuid.UUID, start, end time.Time) (*EngagementMetrics, error) {
	var result struct {
		SessionCount     int     `json:"session_count"`
		TotalTimeMinutes float64 `json:"total_time_minutes"`
	}

	err := h.db.Table("sessions").
		Select(`
			COUNT(*) as session_count,
			COALESCE(SUM(duration_seconds), 0) / 60.0 as total_time_minutes
		`).
		Where("user_id = ?", studentID).
		Where("start_time BETWEEN ? AND ?", start, end).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	averageSessionDuration := 0.0
	if result.SessionCount > 0 {
		averageSessionDuration = result.TotalTimeMinutes / float64(result.SessionCount)
	}

	return &EngagementMetrics{
		SessionCount:           result.SessionCount,
		TotalTimeMinutes:       result.TotalTimeMinutes,
		AverageSessionDuration: averageSessionDuration,
	}, nil
}

func (h *ReportHandler) GetClassroomEngagement(c *gin.Context) {
	classroomID := c.Param("id")
	id, err := uuid.Parse(classroomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid classroom_id format",
			},
		})
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid date format. Use YYYY-MM-DD",
			},
		})
		return
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get active students count
	var activeStudents int64
	h.db.Table("sessions").
		Distinct("user_id").
		Where("classroom_id = ?", id).
		Where("start_time BETWEEN ? AND ?", startOfDay, endOfDay).
		Count(&activeStudents)

	// Get quiz participation metrics
	var quizMetrics struct {
		TotalQuizzes       int `json:"total_quizzes"`
		ParticipatingUsers int `json:"participating_users"`
		TotalResponses     int `json:"total_responses"`
	}

	h.db.Table("quizzes").
		Select(`
			COUNT(DISTINCT quizzes.id) as total_quizzes,
			COUNT(DISTINCT quiz_responses.student_id) as participating_users,
			COUNT(quiz_responses.id) as total_responses
		`).
		Joins("LEFT JOIN quiz_responses ON quizzes.id = quiz_responses.quiz_id").
		Where("quizzes.classroom_id = ?", id).
		Where("quizzes.created_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Scan(&quizMetrics)

	// Calculate participation rate
	participationRate := 0.0
	if activeStudents > 0 {
		participationRate = float64(quizMetrics.ParticipatingUsers) / float64(activeStudents) * 100
	}

	// Get average response time
	var avgResponseTime sql.NullFloat64
	h.db.Table("quiz_responses").
		Select("AVG(time_taken_seconds)").
		Joins("JOIN quizzes ON quiz_responses.quiz_id = quizzes.id").
		Where("quizzes.classroom_id = ?", id).
		Where("quiz_responses.submitted_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Scan(&avgResponseTime)

	responseTime := 0.0
	if avgResponseTime.Valid {
		responseTime = avgResponseTime.Float64
	}

	// Calculate engagement score (simplified)
	engagementScore := (participationRate + float64(quizMetrics.TotalResponses)*2) / 3

	// Get total interactions
	var totalInteractions int64
	h.db.Table("events").
		Where("classroom_id = ?", id).
		Where("timestamp BETWEEN ? AND ?", startOfDay, endOfDay).
		Count(&totalInteractions)

	report := ClassroomEngagementReport{
		ClassroomID: classroomID,
		Date:        date,
		Metrics: ClassroomEngagementMetrics{
			ActiveStudents:        int(activeStudents),
			QuizParticipationRate: participationRate,
			AverageResponseTime:   responseTime,
			EngagementScore:       engagementScore,
			TotalInteractions:     int(totalInteractions),
		},
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) GetClassroomTrends(c *gin.Context) {
	classroomID := c.Param("id")
	id, err := uuid.Parse(classroomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid classroom_id format",
			},
		})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	var trends []models.ClassroomAnalytics
	err = h.db.Table("classroom_analytics").
		Where("classroom_id = ?", id).
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date ASC").
		Scan(&trends).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve trends",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"classroom_id": classroomID,
		"period": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"trends": trends,
	})
}

func (h *ReportHandler) GetContentEffectiveness(c *gin.Context) {
	contentType := c.Query("content_type")
	classroomID := c.Query("classroom_id")
	timePeriod := c.DefaultQuery("time_period", "month")

	// Calculate date range based on time period
	endDate := time.Now()
	var startDate time.Time

	switch timePeriod {
	case "day":
		startDate = endDate.AddDate(0, 0, -1)
	case "week":
		startDate = endDate.AddDate(0, 0, -7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "quarter":
		startDate = endDate.AddDate(0, -3, 0)
	default:
		startDate = endDate.AddDate(0, -1, 0)
	}

	query := h.db.Table("quizzes").
		Select(`
			quizzes.title,
			COUNT(DISTINCT quiz_responses.student_id) as participants,
			AVG(CASE WHEN quiz_responses.points_earned IS NOT NULL THEN
				(quiz_responses.points_earned / quiz_questions.points) * 100
			END) as average_score,
			AVG(quiz_responses.time_taken_seconds) as average_time
		`).
		Joins("LEFT JOIN quiz_responses ON quizzes.id = quiz_responses.quiz_id").
		Joins("LEFT JOIN quiz_questions ON quiz_responses.question_id = quiz_questions.id").
		Where("quizzes.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("quizzes.id, quizzes.title")

	if classroomID != "" {
		if id, err := uuid.Parse(classroomID); err == nil {
			query = query.Where("quizzes.classroom_id = ?", id)
		}
	}

	var effectiveness []map[string]interface{}
	if err := query.Scan(&effectiveness).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to calculate content effectiveness",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content_type": contentType,
		"time_period":  timePeriod,
		"period": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"effectiveness_data": effectiveness,
	})
}