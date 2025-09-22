package seedutils

import (
	"fmt"
	"math/rand"
	"time"

	"reporting-framework/internal/models"

	"gorm.io/gorm"
)

// QuizGenerator handles the creation of realistic quiz data
type QuizGenerator struct {
	db     *gorm.DB
	logger Logger
	rand   *rand.Rand
}

// NewQuizGenerator creates a new quiz generator instance
func NewQuizGenerator(db *gorm.DB, logger Logger) *QuizGenerator {
	source := rand.NewSource(time.Now().UnixNano() + 2000) // Unique offset
	return &QuizGenerator{
		db:     db,
		logger: logger,
		rand:   rand.New(source),
	}
}

// QuizTemplate defines the structure for creating subject-specific quizzes
type QuizTemplate struct {
	SubjectPattern string
	Questions      []QuestionTemplate
	TimeLimit      int
	TotalPoints    float64
}

// QuestionTemplate defines question patterns for different subjects
type QuestionTemplate struct {
	Text    string
	Type    string
	Options map[string]string
	Answer  string
	Points  float64
}

// GetQuizTemplatesBySubject returns realistic quiz templates for different academic subjects
func (qg *QuizGenerator) GetQuizTemplatesBySubject() map[string][]QuizTemplate {
	return map[string][]QuizTemplate{
		"Mathematics": {
			{
				SubjectPattern: "Basic Arithmetic",
				TimeLimit:      20,
				TotalPoints:    100.0,
				Questions: []QuestionTemplate{
					{
						Text:    "What is 15 + 27?",
						Type:    "multiple_choice",
						Options: map[string]string{"A": "42", "B": "41", "C": "43", "D": "40"},
						Answer:  "A",
						Points:  20.0,
					},
					{
						Text:    "What is 8 × 7?",
						Type:    "multiple_choice",
						Options: map[string]string{"A": "54", "B": "56", "C": "58", "D": "52"},
						Answer:  "B",
						Points:  20.0,
					},
				},
			},
			{
				SubjectPattern: "Fractions and Decimals",
				TimeLimit:      25,
				TotalPoints:    100.0,
				Questions: []QuestionTemplate{
					{
						Text:    "What is 1/2 + 1/4?",
						Type:    "multiple_choice",
						Options: map[string]string{"A": "3/4", "B": "2/6", "C": "1/3", "D": "2/4"},
						Answer:  "A",
						Points:  25.0,
					},
				},
			},
		},
		"Science": {
			{
				SubjectPattern: "Basic Biology",
				TimeLimit:      30,
				TotalPoints:    100.0,
				Questions: []QuestionTemplate{
					{
						Text:    "What part of the plant conducts photosynthesis?",
						Type:    "multiple_choice",
						Options: map[string]string{"A": "Roots", "B": "Leaves", "C": "Stem", "D": "Flowers"},
						Answer:  "B",
						Points:  25.0,
					},
				},
			},
		},
		"English Language Arts": {
			{
				SubjectPattern: "Grammar and Vocabulary",
				TimeLimit:      25,
				TotalPoints:    100.0,
				Questions: []QuestionTemplate{
					{
						Text:    "Which word is a noun in this sentence: 'The cat ran quickly'?",
						Type:    "multiple_choice",
						Options: map[string]string{"A": "The", "B": "cat", "C": "ran", "D": "quickly"},
						Answer:  "B",
						Points:  25.0,
					},
				},
			},
		},
		"Social Studies": {
			{
				SubjectPattern: "American History Basics",
				TimeLimit:      20,
				TotalPoints:    100.0,
				Questions: []QuestionTemplate{
					{
						Text:    "Who was the first President of the United States?",
						Type:    "multiple_choice",
						Options: map[string]string{"A": "Thomas Jefferson", "B": "George Washington", "C": "John Adams", "D": "Benjamin Franklin"},
						Answer:  "B",
						Points:  25.0,
					},
				},
			},
		},
	}
}

// GenerateQuizzesForClassrooms creates realistic quizzes for each classroom based on subject
func (qg *QuizGenerator) GenerateQuizzesForClassrooms(classrooms []models.Classroom, quizzesPerClassroom int) ([]models.Quiz, error) {
	qg.logger.Info("Starting quiz generation", "classrooms", len(classrooms), "perClassroom", quizzesPerClassroom)

	if len(classrooms) == 0 {
		return nil, fmt.Errorf("cannot generate quizzes: no classrooms provided")
	}

	templates := qg.GetQuizTemplatesBySubject()
	var allQuizzes []models.Quiz

	for _, classroom := range classrooms {
		subjectTemplates, exists := templates[classroom.Subject]
		if !exists {
			// Create generic template for subjects not in our predefined list
			subjectTemplates = []QuizTemplate{
				{
					SubjectPattern: fmt.Sprintf("%s Fundamentals", classroom.Subject),
					TimeLimit:      30,
					TotalPoints:    100.0,
					Questions: []QuestionTemplate{
						{
							Text:    fmt.Sprintf("Basic %s question", classroom.Subject),
							Type:    "multiple_choice",
							Options: map[string]string{"A": "Option A", "B": "Option B", "C": "Option C", "D": "Option D"},
							Answer:  "A",
							Points:  25.0,
						},
					},
				},
			}
		}

		classroomQuizzes, err := qg.generateQuizzesForSingleClassroom(classroom, subjectTemplates, quizzesPerClassroom)
		if err != nil {
			return nil, fmt.Errorf("failed to generate quizzes for classroom %s: %w", classroom.Name, err)
		}

		allQuizzes = append(allQuizzes, classroomQuizzes...)
	}

	qg.logger.Info("Successfully generated all quizzes", "total", len(allQuizzes))
	return allQuizzes, nil
}

// generateQuizzesForSingleClassroom creates quizzes for a specific classroom
func (qg *QuizGenerator) generateQuizzesForSingleClassroom(classroom models.Classroom, templates []QuizTemplate, count int) ([]models.Quiz, error) {
	var quizzes []models.Quiz

	for i := 0; i < count; i++ {
		template := templates[i%len(templates)]

		// Create realistic quiz timing - some recent, some older
		daysAgo := qg.rand.Intn(60) // 0-60 days ago
		publishedAt := time.Now().AddDate(0, 0, -daysAgo)

		quiz := models.Quiz{
			Title:            fmt.Sprintf("%s Quiz #%d - %s", classroom.Subject, i+1, template.SubjectPattern),
			ClassroomID:      classroom.ID,
			TeacherID:        classroom.TeacherID,
			QuestionCount:    len(template.Questions),
			TotalPoints:      template.TotalPoints,
			TimeLimitMinutes: &template.TimeLimit,
			Status:           "published",
			PublishedAt:      &publishedAt,
		}

		if err := qg.db.Create(&quiz).Error; err != nil {
			return nil, fmt.Errorf("failed to create quiz %s: %w", quiz.Title, err)
		}

		// Create questions for this quiz
		if err := qg.createQuestionsForQuiz(quiz, template.Questions); err != nil {
			return nil, fmt.Errorf("failed to create questions for quiz %s: %w", quiz.Title, err)
		}

		quizzes = append(quizzes, quiz)
	}

	qg.logger.Debug("Created quizzes for classroom", "classroom", classroom.Name, "count", len(quizzes))
	return quizzes, nil
}

// createQuestionsForQuiz creates questions based on templates
func (qg *QuizGenerator) createQuestionsForQuiz(quiz models.Quiz, questionTemplates []QuestionTemplate) error {
	for i, template := range questionTemplates {
		// Convert string map to JSONB
		optionsJSONB := make(models.JSONB)
		for k, v := range template.Options {
			optionsJSONB[k] = v
		}

		question := models.QuizQuestion{
			QuizID:        quiz.ID,
			QuestionText:  template.Text,
			QuestionType:  template.Type,
			Options:       optionsJSONB,
			CorrectAnswer: template.Answer,
			Points:        template.Points,
			OrderIndex:    i + 1,
		}

		if err := qg.db.Create(&question).Error; err != nil {
			return fmt.Errorf("failed to create question %d for quiz %s: %w", i+1, quiz.Title, err)
		}
	}

	return nil
}

// ResponseGenerator handles creation of realistic quiz responses
type ResponseGenerator struct {
	db     *gorm.DB
	logger Logger
	rand   *rand.Rand
}

// NewResponseGenerator creates a new response generator
func NewResponseGenerator(db *gorm.DB, logger Logger) *ResponseGenerator {
	source := rand.NewSource(time.Now().UnixNano() + 3000)
	return &ResponseGenerator{
		db:     db,
		logger: logger,
		rand:   rand.New(source),
	}
}

// ResponsePattern defines realistic student response patterns
type ResponsePattern struct {
	ParticipationRate float64 // Percentage of students who participate
	CorrectAnswerRate float64 // Percentage of correct answers
	MinResponseTime   int     // Minimum response time in seconds
	MaxResponseTime   int     // Maximum response time in seconds
}

// GetRealisticResponsePattern returns patterns that simulate real classroom behavior
func GetRealisticResponsePattern() ResponsePattern {
	return ResponsePattern{
		ParticipationRate: 0.85, // 85% of students typically participate
		CorrectAnswerRate: 0.72, // 72% accuracy rate is realistic for K-8
		MinResponseTime:   15,   // Minimum 15 seconds thinking time
		MaxResponseTime:   180,  // Maximum 3 minutes per question
	}
}

// GenerateResponsesForAllQuizzes creates realistic student responses
func (rg *ResponseGenerator) GenerateResponsesForAllQuizzes(quizzes []models.Quiz) error {
	rg.logger.Info("Starting response generation for all quizzes", "quizCount", len(quizzes))

	pattern := GetRealisticResponsePattern()
	totalResponses := 0

	for i, quiz := range quizzes {
		responses, err := rg.generateResponsesForQuiz(quiz, pattern)
		if err != nil {
			return fmt.Errorf("failed to generate responses for quiz %s: %w", quiz.Title, err)
		}

		totalResponses += responses

		if (i+1)%20 == 0 { // Log every 20 quizzes to avoid spam
			rg.logger.Debug("Quiz response generation progress", "processed", i+1, "total", len(quizzes), "responses", totalResponses)
		}
	}

	rg.logger.Info("Successfully generated all quiz responses", "totalResponses", totalResponses)
	return nil
}

// generateResponsesForQuiz creates responses for a single quiz
func (rg *ResponseGenerator) generateResponsesForQuiz(quiz models.Quiz, pattern ResponsePattern) (int, error) {
	// Get enrolled students for this classroom
	var enrolledStudents []models.User
	err := rg.db.Table("users").
		Joins("JOIN enrollments ON users.id = enrollments.user_id").
		Where("enrollments.classroom_id = ? AND users.role = 'student'", quiz.ClassroomID).
		Find(&enrolledStudents).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get enrolled students: %w", err)
	}

	// Get questions for this quiz
	var questions []models.QuizQuestion
	err = rg.db.Where("quiz_id = ?", quiz.ID).Order("order_index").Find(&questions).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get quiz questions: %w", err)
	}

	responseCount := 0

	for _, student := range enrolledStudents {
		// Simulate participation rate
		if rg.rand.Float64() > pattern.ParticipationRate {
			continue // Student didn't participate
		}

		for _, question := range questions {
			// Determine if answer is correct based on pattern
			isCorrect := rg.rand.Float64() < pattern.CorrectAnswerRate

			answer := question.CorrectAnswer
			if !isCorrect {
				// Generate a wrong answer
				answer = rg.generateWrongAnswer(question)
			}

			pointsEarned := 0.0
			if isCorrect {
				pointsEarned = question.Points
			}

			// Generate realistic response time
			timeTaken := pattern.MinResponseTime + rg.rand.Intn(pattern.MaxResponseTime-pattern.MinResponseTime)

			// Generate submission time relative to quiz publication
			var submittedAt time.Time
			if quiz.PublishedAt != nil {
				// Response submitted within a week of quiz publication
				maxDays := 7
				dayOffset := rg.rand.Intn(maxDays)
				hourOffset := rg.rand.Intn(24)
				submittedAt = quiz.PublishedAt.AddDate(0, 0, dayOffset).Add(time.Duration(hourOffset) * time.Hour)
			} else {
				submittedAt = time.Now().AddDate(0, 0, -rg.rand.Intn(30))
			}

			response := models.QuizResponse{
				QuizID:           quiz.ID,
				QuestionID:       question.ID,
				StudentID:        student.ID,
				Answer:           answer,
				IsCorrect:        &isCorrect,
				PointsEarned:     &pointsEarned,
				TimeTakenSeconds: &timeTaken,
				SubmittedAt:      &submittedAt,
			}

			if err := rg.db.Create(&response).Error; err != nil {
				return responseCount, fmt.Errorf("failed to create response for student %s: %w", student.Username, err)
			}

			responseCount++
		}
	}

	return responseCount, nil
}

// generateWrongAnswer creates a plausible incorrect answer
func (rg *ResponseGenerator) generateWrongAnswer(question models.QuizQuestion) string {
	// Convert options back from JSONB
	options := question.Options
	var wrongAnswers []string

	for key := range options {
		if key != question.CorrectAnswer {
			wrongAnswers = append(wrongAnswers, key)
		}
	}

	if len(wrongAnswers) == 0 {
		// Fallback for questions without multiple choice options
		return "incorrect_answer"
	}

	return wrongAnswers[rg.rand.Intn(len(wrongAnswers))]
}