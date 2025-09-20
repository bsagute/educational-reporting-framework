package handlers

import (
	"net/http"
	"time"

	"reporting-framework/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizHandler struct {
	db *gorm.DB
}

type CreateQuizRequest struct {
	Title            string    `json:"title" binding:"required"`
	ClassroomID      string    `json:"classroom_id" binding:"required"`
	TeacherID        string    `json:"teacher_id" binding:"required"`
	TimeLimitMinutes *int      `json:"time_limit_minutes"`
	Questions        []Question `json:"questions"`
}

type Question struct {
	QuestionText  string                 `json:"question_text" binding:"required"`
	QuestionType  string                 `json:"question_type" binding:"required"`
	Options       map[string]interface{} `json:"options"`
	CorrectAnswer string                 `json:"correct_answer"`
	Points        float64                `json:"points"`
	OrderIndex    int                    `json:"order_index"`
}

type SubmitResponseRequest struct {
	QuestionID       string `json:"question_id" binding:"required"`
	StudentID        string `json:"student_id" binding:"required"`
	Answer           string `json:"answer" binding:"required"`
	TimeTakenSeconds int    `json:"time_taken_seconds"`
}

func NewQuizHandler(db *gorm.DB) *QuizHandler {
	return &QuizHandler{db: db}
}

func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var req CreateQuizRequest
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

	classroomID, err := uuid.Parse(req.ClassroomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid classroom_id format",
			},
		})
		return
	}

	teacherID, err := uuid.Parse(req.TeacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid teacher_id format",
			},
		})
		return
	}

	// Start transaction
	tx := h.db.Begin()

	// Create quiz
	quiz := models.Quiz{
		Title:            req.Title,
		ClassroomID:      classroomID,
		TeacherID:        teacherID,
		QuestionCount:    len(req.Questions),
		TimeLimitMinutes: req.TimeLimitMinutes,
		Status:           "draft",
	}

	if err := tx.Create(&quiz).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create quiz",
				"details": err.Error(),
			},
		})
		return
	}

	// Create questions
	var totalPoints float64
	for _, questionData := range req.Questions {
		question := models.QuizQuestion{
			QuizID:        quiz.ID,
			QuestionText:  questionData.QuestionText,
			QuestionType:  questionData.QuestionType,
			Options:       models.JSONB(questionData.Options),
			CorrectAnswer: questionData.CorrectAnswer,
			Points:        questionData.Points,
			OrderIndex:    questionData.OrderIndex,
		}

		if err := tx.Create(&question).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]interface{}{
					"code":    "DATABASE_ERROR",
					"message": "Failed to create question",
					"details": err.Error(),
				},
			})
			return
		}

		totalPoints += questionData.Points
	}

	// Update quiz with total points
	if err := tx.Model(&quiz).Update("total_points", totalPoints).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update quiz total points",
				"details": err.Error(),
			},
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"quiz_id":      quiz.ID,
		"question_count": quiz.QuestionCount,
		"total_points": totalPoints,
	})
}

func (h *QuizHandler) UpdateQuiz(c *gin.Context) {
	quizID := c.Param("id")
	id, err := uuid.Parse(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid quiz_id format",
			},
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
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

	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.Status == "published" {
		updates["published_at"] = time.Now()
	}

	if err := h.db.Model(&models.Quiz{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update quiz",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Quiz updated successfully",
	})
}

func (h *QuizHandler) SubmitResponse(c *gin.Context) {
	quizID := c.Param("id")
	id, err := uuid.Parse(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid quiz_id format",
			},
		})
		return
	}

	var req SubmitResponseRequest
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

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid question_id format",
			},
		})
		return
	}

	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid student_id format",
			},
		})
		return
	}

	// Get question to check correct answer
	var question models.QuizQuestion
	if err := h.db.First(&question, "id = ?", questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": map[string]interface{}{
				"code":    "NOT_FOUND",
				"message": "Question not found",
			},
		})
		return
	}

	// Check if answer is correct
	isCorrect := req.Answer == question.CorrectAnswer
	var pointsEarned float64
	if isCorrect {
		pointsEarned = question.Points
	}

	response := models.QuizResponse{
		QuizID:           id,
		QuestionID:       questionID,
		StudentID:        studentID,
		Answer:           req.Answer,
		IsCorrect:        &isCorrect,
		PointsEarned:     &pointsEarned,
		TimeTakenSeconds: &req.TimeTakenSeconds,
		SubmittedAt:      &time.Time{},
	}
	*response.SubmittedAt = time.Now()

	if err := h.db.Create(&response).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to submit response",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"is_correct":     isCorrect,
		"points_earned":  pointsEarned,
		"response_id":    response.ID,
	})
}

func (h *QuizHandler) GetQuiz(c *gin.Context) {
	quizID := c.Param("id")
	id, err := uuid.Parse(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid quiz_id format",
			},
		})
		return
	}

	var quiz models.Quiz
	if err := h.db.Preload("Teacher").Preload("Classroom").First(&quiz, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "Quiz not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve quiz",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, quiz)
}