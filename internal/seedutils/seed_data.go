package seedutils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"reporting-framework/internal/domain/reporting"
)

// SeedManager handles seeding the database with test data
type SeedManager struct {
	db *gorm.DB
}

// NewSeedManager creates a new seed manager
func NewSeedManager(db *gorm.DB) *SeedManager {
	return &SeedManager{db: db}
}

// SeedAllData seeds the database with comprehensive test data
func (s *SeedManager) SeedAllData() error {
	fmt.Println("Starting database seeding...")

	// 1. Create schools
	schools, err := s.seedSchools(10) // 10 schools for testing
	if err != nil {
		return fmt.Errorf("failed to seed schools: %w", err)
	}

	// 2. Create classrooms (30 per school)
	classrooms, err := s.seedClassrooms(schools, 30)
	if err != nil {
		return fmt.Errorf("failed to seed classrooms: %w", err)
	}

	// 3. Create users (teachers and students)
	users, err := s.seedUsers(schools, classrooms)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// 4. Create quizzes
	quizzes, err := s.seedQuizzes(classrooms, users)
	if err != nil {
		return fmt.Errorf("failed to seed quizzes: %w", err)
	}

	// 5. Create sessions and events (historical data)
	err = s.seedSessionsAndEvents(users, classrooms, 30) // 30 days of historical data
	if err != nil {
		return fmt.Errorf("failed to seed sessions and events: %w", err)
	}

	// 6. Create quiz sessions and submissions
	err = s.seedQuizData(quizzes, users)
	if err != nil {
		return fmt.Errorf("failed to seed quiz data: %w", err)
	}

	// 7. Create content
	err = s.seedContent(users, classrooms)
	if err != nil {
		return fmt.Errorf("failed to seed content: %w", err)
	}

	// 8. Generate aggregated metrics
	err = s.generateAggregatedMetrics()
	if err != nil {
		return fmt.Errorf("failed to generate aggregated metrics: %w", err)
	}

	fmt.Println("Database seeding completed successfully!")
	return nil
}

// seedSchools creates test schools
func (s *SeedManager) seedSchools(count int) ([]reporting.School, error) {
	schools := make([]reporting.School, count)

	schoolNames := []string{
		"Lincoln Elementary", "Washington High School", "Roosevelt Middle School",
		"Jefferson Academy", "Madison School", "Hamilton Institute",
		"Franklin School", "Adams Elementary", "Wilson High", "Monroe Academy",
	}

	districts := []string{"North District", "South District", "East District", "West District", "Central District"}
	regions := []string{"North Region", "South Region", "East Region", "West Region"}

	for i := 0; i < count; i++ {
		school := reporting.School{
			ID:       uuid.New(),
			Name:     schoolNames[i%len(schoolNames)] + fmt.Sprintf(" #%d", i+1),
			District: &districts[i%len(districts)],
			Region:   &regions[i%len(regions)],
		}
		email := fmt.Sprintf("admin@school%d.edu", i+1)
		school.ContactEmail = &email

		schools[i] = school
	}

	return schools, s.db.CreateInBatches(schools, 100).Error
}

// seedClassrooms creates test classrooms
func (s *SeedManager) seedClassrooms(schools []reporting.School, classroomsPerSchool int) ([]reporting.Classroom, error) {
	var classrooms []reporting.Classroom

	subjects := []string{"Mathematics", "Science", "English", "History", "Art", "Physical Education", "Music"}
	gradeLevels := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	for _, school := range schools {
		for i := 0; i < classroomsPerSchool; i++ {
			classroom := reporting.Classroom{
				ID:         uuid.New(),
				SchoolID:   school.ID,
				Name:       fmt.Sprintf("Room %d%02d", (i/10)+1, (i%10)+1),
				Subject:    &subjects[i%len(subjects)],
				MaxStudents: 30,
			}

			gradeLevel := gradeLevels[i%len(gradeLevels)]
			classroom.GradeLevel = &gradeLevel

			classrooms = append(classrooms, classroom)
		}
	}

	return classrooms, s.db.CreateInBatches(classrooms, 100).Error
}

// seedUsers creates test users (teachers and students)
func (s *SeedManager) seedUsers(schools []reporting.School, classrooms []reporting.Classroom) ([]reporting.User, error) {
	var users []reporting.User
	var userClassrooms []reporting.UserClassroom

	// Create teachers (1 per classroom)
	for _, classroom := range classrooms {
		teacher := reporting.User{
			ID:       uuid.New(),
			SchoolID: classroom.SchoolID,
			Username: fmt.Sprintf("teacher_%s", classroom.ID.String()[:8]),
			Role:     "teacher",
		}

		firstName := fmt.Sprintf("Teacher%d", len(users)+1)
		lastName := "Smith"
		email := fmt.Sprintf("%s@school.edu", teacher.Username)

		teacher.FirstName = &firstName
		teacher.LastName = &lastName
		teacher.Email = &email
		teacher.LastActive = &time.Time{}
		*teacher.LastActive = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)

		users = append(users, teacher)

		// Assign teacher to classroom
		userClassrooms = append(userClassrooms, reporting.UserClassroom{
			UserID:      teacher.ID,
			ClassroomID: classroom.ID,
			Role:        "teacher",
			IsActive:    true,
		})

		// Update classroom with teacher ID
		s.db.Model(&classroom).Update("teacher_id", teacher.ID)
	}

	// Create students (30 per classroom)
	studentCount := 0
	for _, classroom := range classrooms {
		for i := 0; i < 30; i++ {
			studentCount++
			student := reporting.User{
				ID:       uuid.New(),
				SchoolID: classroom.SchoolID,
				Username: fmt.Sprintf("student_%d", studentCount),
				Role:     "student",
			}

			firstName := fmt.Sprintf("Student%d", studentCount)
			lastName := "Johnson"
			email := fmt.Sprintf("%s@school.edu", student.Username)

			student.FirstName = &firstName
			student.LastName = &lastName
			student.Email = &email
			student.LastActive = &time.Time{}
			*student.LastActive = time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour)

			users = append(users, student)

			// Assign student to classroom
			userClassrooms = append(userClassrooms, reporting.UserClassroom{
				UserID:      student.ID,
				ClassroomID: classroom.ID,
				Role:        "student",
				IsActive:    true,
			})
		}
	}

	// Batch insert users
	if err := s.db.CreateInBatches(users, 100).Error; err != nil {
		return nil, err
	}

	// Batch insert user-classroom relationships
	return users, s.db.CreateInBatches(userClassrooms, 100).Error
}

// seedQuizzes creates test quizzes
func (s *SeedManager) seedQuizzes(classrooms []reporting.Classroom, users []reporting.User) ([]reporting.Quiz, error) {
	var quizzes []reporting.Quiz
	var questions []reporting.QuizQuestion

	// Get teachers
	teachers := make(map[uuid.UUID]reporting.User)
	for _, user := range users {
		if user.Role == "teacher" {
			teachers[user.ID] = user
		}
	}

	quizTitles := []string{
		"Math Quiz #1", "Science Quiz #1", "History Quiz #1", "English Quiz #1",
		"Weekly Assessment", "Chapter Review", "Unit Test", "Pop Quiz",
		"Midterm Exam", "Final Exam", "Practice Test", "Skills Assessment",
	}

	for _, classroom := range classrooms {
		if classroom.TeacherID == nil {
			continue
		}

		// Create 3-5 quizzes per classroom
		numQuizzes := 3 + rand.Intn(3)
		for i := 0; i < numQuizzes; i++ {
			quiz := reporting.Quiz{
				ID:          uuid.New(),
				CreatorID:   *classroom.TeacherID,
				ClassroomID: classroom.ID,
				Title:       quizTitles[rand.Intn(len(quizTitles))],
				MaxAttempts: 1 + rand.Intn(3),
				IsActive:    rand.Float32() < 0.3, // 30% chance of being active
			}

			description := fmt.Sprintf("Assessment for %s classroom", *classroom.Subject)
			quiz.Description = &description

			if quiz.IsActive {
				startTime := time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
				endTime := startTime.Add(time.Duration(60+rand.Intn(120)) * time.Minute)
				quiz.StartTime = &startTime
				quiz.EndTime = &endTime
			}

			timeLimit := 30 + rand.Intn(90)
			quiz.TimeLimitMinutes = &timeLimit

			quizzes = append(quizzes, quiz)

			// Create 5-10 questions per quiz
			numQuestions := 5 + rand.Intn(6)
			quiz.TotalQuestions = numQuestions

			for j := 0; j < numQuestions; j++ {
				question := reporting.QuizQuestion{
					ID:           uuid.New(),
					QuizID:       quiz.ID,
					QuestionText: fmt.Sprintf("Question %d for %s", j+1, quiz.Title),
					QuestionType: []string{"multiple_choice", "true_false", "short_answer"}[rand.Intn(3)],
					Points:       1 + rand.Intn(5),
					OrderIndex:   j + 1,
				}

				if question.QuestionType == "multiple_choice" {
					options := map[string]interface{}{
						"A": "Option A",
						"B": "Option B",
						"C": "Option C",
						"D": "Option D",
					}
					question.Options = reporting.JSONB(options)
					answers := []string{"A", "B", "C", "D"}
					correctAnswer := answers[rand.Intn(len(answers))]
					question.CorrectAnswer = &correctAnswer
				} else if question.QuestionType == "true_false" {
					answers := []string{"true", "false"}
					correctAnswer := answers[rand.Intn(len(answers))]
					question.CorrectAnswer = &correctAnswer
				}

				questions = append(questions, question)
			}
		}
	}

	// Insert quizzes
	if err := s.db.CreateInBatches(quizzes, 100).Error; err != nil {
		return nil, err
	}

	// Insert questions
	if err := s.db.CreateInBatches(questions, 100).Error; err != nil {
		return nil, err
	}

	return quizzes, nil
}

// seedSessionsAndEvents creates historical sessions and events
func (s *SeedManager) seedSessionsAndEvents(users []reporting.User, classrooms []reporting.Classroom, days int) error {
	var sessions []reporting.Session
	var events []reporting.Event

	applications := []string{"whiteboard", "notebook"}
	eventTypes := []string{
		"session_start", "session_end", "content_created", "content_viewed",
		"whiteboard_draw", "notebook_write", "quiz_started", "quiz_completed",
		"content_shared", "sync_initiated",
	}

	for d := 0; d < days; d++ {
		date := time.Now().AddDate(0, 0, -days+d)

		// Simulate 60-80% of users being active each day
		activeUsers := users[0:int(float64(len(users)) * (0.6 + rand.Float64()*0.2))]

		for _, user := range activeUsers {
			// Each active user has 1-3 sessions per day
			numSessions := 1 + rand.Intn(3)

			for i := 0; i < numSessions; i++ {
				session := reporting.Session{
					ID:          uuid.New(),
					UserID:      user.ID,
					Application: applications[rand.Intn(len(applications))],
					StartTime:   date.Add(time.Duration(8+rand.Intn(10)) * time.Hour),
				}

				// Assign classroom based on user role
				if user.Role == "student" {
					// Students are in their assigned classroom
					var userClassroom reporting.UserClassroom
					s.db.Where("user_id = ? AND role = 'student'", user.ID).First(&userClassroom)
					session.ClassroomID = &userClassroom.ClassroomID
				} else if user.Role == "teacher" {
					// Teachers can be in any of their classrooms
					var userClassroom reporting.UserClassroom
					s.db.Where("user_id = ? AND role = 'teacher'", user.ID).First(&userClassroom)
					session.ClassroomID = &userClassroom.ClassroomID
				}

				// Session duration: 15-120 minutes
				duration := 15 + rand.Intn(105)
				endTime := session.StartTime.Add(time.Duration(duration) * time.Minute)
				session.EndTime = &endTime
				session.DurationSeconds = &[]int{duration * 60}[0]

				// Device info
				deviceInfo := map[string]interface{}{
					"platform":    "android",
					"version":     "10.0",
					"device_model": "Samsung Galaxy Tab",
				}
				session.DeviceInfo = reporting.JSONB(deviceInfo)

				sessions = append(sessions, session)

				// Generate 5-20 events per session
				numEvents := 5 + rand.Intn(16)
				for j := 0; j < numEvents; j++ {
					event := reporting.Event{
						ID:          uuid.New(),
						EventType:   eventTypes[rand.Intn(len(eventTypes))],
						UserID:      &user.ID,
						SessionID:   &session.ID,
						ClassroomID: session.ClassroomID,
						SchoolID:    &user.SchoolID,
						Application: &session.Application,
						Timestamp:   session.StartTime.Add(time.Duration(j*duration/numEvents) * time.Minute),
					}

					// Event metadata
					metadata := map[string]interface{}{
						"duration": rand.Intn(300),
						"action_count": rand.Intn(50),
					}
					event.Metadata = reporting.JSONB(metadata)
					event.DeviceInfo = session.DeviceInfo

					events = append(events, event)
				}
			}
		}
	}

	// Batch insert sessions
	if err := s.db.CreateInBatches(sessions, 100).Error; err != nil {
		return err
	}

	// Batch insert events
	return s.db.CreateInBatches(events, 100).Error
}

// seedQuizData creates quiz sessions and submissions
func (s *SeedManager) seedQuizData(quizzes []reporting.Quiz, users []reporting.User) error {
	var quizSessions []reporting.QuizSession
	var submissions []reporting.QuizSubmission

	// Get students by classroom
	studentsByClassroom := make(map[uuid.UUID][]reporting.User)
	for _, user := range users {
		if user.Role == "student" {
			var userClassroom reporting.UserClassroom
			s.db.Where("user_id = ? AND role = 'student'", user.ID).First(&userClassroom)
			studentsByClassroom[userClassroom.ClassroomID] = append(
				studentsByClassroom[userClassroom.ClassroomID], user)
		}
	}

	for _, quiz := range quizzes {
		students := studentsByClassroom[quiz.ClassroomID]
		if len(students) == 0 {
			continue
		}

		// 60-90% of students participate in each quiz
		participationRate := 0.6 + rand.Float64()*0.3
		participatingStudents := students[0:int(float64(len(students)) * participationRate)]

		// Get quiz questions
		var questions []reporting.QuizQuestion
		s.db.Where("quiz_id = ?", quiz.ID).Order("order_index").Find(&questions)

		for _, student := range participatingStudents {
			// Create quiz session
			session := reporting.QuizSession{
				ID:        uuid.New(),
				QuizID:    quiz.ID,
				StudentID: student.ID,
				StartedAt: time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour),
			}

			totalScore := 0
			maxScore := 0

			// Create submissions for each question
			for _, question := range questions {
				submission := reporting.QuizSubmission{
					ID:         uuid.New(),
					QuizID:     quiz.ID,
					StudentID:  student.ID,
					QuestionID: question.ID,
					TimeSpentSeconds: func() *int { v := 30 + rand.Intn(120); return &v }(),
				}

				maxScore += question.Points

				// Simulate answer correctness (70% correct on average)
				isCorrect := rand.Float32() < 0.7
				submission.IsCorrect = &isCorrect

				if isCorrect {
					totalScore += question.Points
					submission.PointsEarned = question.Points
				}

				// Generate answer based on question type
				if question.QuestionType == "multiple_choice" {
					answers := []string{"A", "B", "C", "D"}
					if isCorrect && question.CorrectAnswer != nil {
						submission.SubmittedAnswer = question.CorrectAnswer
					} else {
						answer := answers[rand.Intn(len(answers))]
						submission.SubmittedAnswer = &answer
					}
				} else if question.QuestionType == "true_false" {
					if isCorrect && question.CorrectAnswer != nil {
						submission.SubmittedAnswer = question.CorrectAnswer
					} else {
						answers := []string{"true", "false"}
						answer := answers[rand.Intn(len(answers))]
						submission.SubmittedAnswer = &answer
					}
				} else {
					answer := "Sample short answer"
					submission.SubmittedAnswer = &answer
				}

				submissions = append(submissions, submission)
			}

			// Complete session
			session.TotalScore = totalScore
			session.MaxPossibleScore = maxScore
			if maxScore > 0 {
				percentage := float64(totalScore) / float64(maxScore) * 100
				session.PercentageScore = &percentage
			}
			session.IsCompleted = true
			completedAt := session.StartedAt.Add(time.Duration(len(questions)*2) * time.Minute)
			session.CompletedAt = &completedAt
			timeSpent := int(completedAt.Sub(session.StartedAt).Seconds())
			session.TimeSpentSeconds = &timeSpent

			quizSessions = append(quizSessions, session)
		}
	}

	// Insert quiz sessions
	if err := s.db.CreateInBatches(quizSessions, 100).Error; err != nil {
		return err
	}

	// Insert submissions
	return s.db.CreateInBatches(submissions, 100).Error
}

// seedContent creates test content
func (s *SeedManager) seedContent(users []reporting.User, classrooms []reporting.Classroom) error {
	var content []reporting.Content

	contentTypes := []string{"note", "drawing", "document", "whiteboard_session"}

	for _, classroom := range classrooms {
		// Get classroom users
		var classroomUsers []reporting.User
		s.db.Table("users u").
			Joins("JOIN user_classrooms uc ON u.id = uc.user_id").
			Where("uc.classroom_id = ?", classroom.ID).
			Find(&classroomUsers)

		// Each user creates 5-15 pieces of content
		for _, user := range classroomUsers {
			numContent := 5 + rand.Intn(11)
			for i := 0; i < numContent; i++ {
				item := reporting.Content{
					ID:          uuid.New(),
					CreatorID:   user.ID,
					ClassroomID: &classroom.ID,
					ContentType: contentTypes[rand.Intn(len(contentTypes))],
					IsShared:    rand.Float32() < 0.3, // 30% shared
					FileSizeBytes: int64(1024 + rand.Intn(1024*1024)), // 1KB to 1MB
				}

				title := fmt.Sprintf("%s Content %d", item.ContentType, i+1)
				item.Title = &title

				// Content data
				contentData := map[string]interface{}{
					"version": "1.0",
					"created_by": user.Username,
					"tags": []string{"education", *classroom.Subject},
				}
				item.ContentData = reporting.JSONB(contentData)

				if item.IsShared {
					sharePermissions := map[string]interface{}{
						"read": []string{"student", "teacher"},
						"write": []string{"teacher"},
					}
					item.SharePermissions = reporting.JSONB(sharePermissions)
				}

				content = append(content, item)
			}
		}
	}

	return s.db.CreateInBatches(content, 100).Error
}

// generateAggregatedMetrics creates initial aggregated metrics
func (s *SeedManager) generateAggregatedMetrics() error {
	// This would typically be done by background jobs
	// For demo purposes, we'll create some sample aggregated data

	fmt.Println("Generating aggregated metrics...")

	// Generate daily user metrics for the past 30 days
	for d := 0; d < 30; d++ {
		date := time.Now().AddDate(0, 0, -30+d)

		// This is a simplified version - in production, this would be calculated from actual events
		err := s.db.Exec(`
			INSERT INTO daily_user_metrics (
				user_id, school_id, date, session_count, total_session_duration_seconds,
				events_count, quiz_attempts, avg_quiz_score, content_created_count
			)
			SELECT
				u.id as user_id,
				u.school_id,
				?::date as date,
				1 + FLOOR(RANDOM() * 3) as session_count,
				(1800 + FLOOR(RANDOM() * 3600))::int as total_session_duration_seconds,
				(10 + FLOOR(RANDOM() * 50))::int as events_count,
				FLOOR(RANDOM() * 3)::int as quiz_attempts,
				(60 + RANDOM() * 35)::decimal(5,2) as avg_quiz_score,
				FLOOR(RANDOM() * 5)::int as content_created_count
			FROM users u
			WHERE u.role IN ('student', 'teacher')
			AND RANDOM() > 0.3  -- 70% of users active each day
			ON CONFLICT (user_id, date) DO NOTHING
		`, date).Error

		if err != nil {
			return err
		}
	}

	// Generate daily classroom metrics
	err := s.db.Exec(`
		INSERT INTO daily_classroom_metrics (
			classroom_id, school_id, date, total_students, active_students_count,
			participation_rate, avg_session_duration_minutes, engagement_score
		)
		SELECT
			c.id as classroom_id,
			c.school_id,
			CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 29) as date,
			30 as total_students,
			(20 + FLOOR(RANDOM() * 10))::int as active_students_count,
			(65 + RANDOM() * 30)::decimal(5,2) as participation_rate,
			(35 + RANDOM() * 20)::decimal(10,2) as avg_session_duration_minutes,
			(70 + RANDOM() * 25)::decimal(5,2) as engagement_score
		FROM classrooms c
		ON CONFLICT (classroom_id, date) DO NOTHING
	`).Error

	return err
}

// CleanupTestData removes all test data from the database
func (s *SeedManager) CleanupTestData() error {
	fmt.Println("Cleaning up test data...")

	tables := []string{
		"daily_classroom_metrics", "daily_user_metrics", "weekly_school_metrics",
		"content_metrics", "quiz_analytics", "active_sessions", "hourly_event_aggregates",
		"events", "quiz_submissions", "quiz_sessions", "quiz_questions", "quizzes",
		"content", "sessions", "user_classrooms", "users", "classrooms", "schools",
	}

	for _, table := range tables {
		if err := s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	fmt.Println("Test data cleanup completed!")
	return nil
}