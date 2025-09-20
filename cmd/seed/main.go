package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"reporting-framework/internal/config"
	"reporting-framework/internal/database"
	"reporting-framework/internal/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations first
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Starting data seeding...")

	// Seed sample data
	if err := seedSampleData(db); err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	log.Println("Data seeding completed successfully!")
}

func seedSampleData(db *gorm.DB) error {
	// Create sample schools
	schools := []models.School{
		{
			Name:     "Lincoln Elementary School",
			District: "Downtown District",
			Region:   "North",
			Timezone: "America/New_York",
		},
		{
			Name:     "Washington Middle School",
			District: "Suburban District",
			Region:   "South",
			Timezone: "America/New_York",
		},
		{
			Name:     "Roosevelt High School",
			District: "Metro District",
			Region:   "East",
			Timezone: "America/New_York",
		},
	}

	for i := range schools {
		if err := db.Create(&schools[i]).Error; err != nil {
			return err
		}
	}

	// Create sample teachers
	teachers := []models.User{}
	for i := 0; i < 15; i++ {
		teacher := models.User{
			Email:     fmt.Sprintf("teacher%d@school.edu", i+1),
			Username:  fmt.Sprintf("teacher%d", i+1),
			FirstName: fmt.Sprintf("Teacher%d", i+1),
			LastName:  "Smith",
			Role:      "teacher",
			SchoolID:  schools[i%len(schools)].ID,
		}
		teachers = append(teachers, teacher)
	}

	for i := range teachers {
		if err := db.Create(&teachers[i]).Error; err != nil {
			return err
		}
	}

	// Create sample classrooms
	subjects := []string{"Mathematics", "Science", "English", "History", "Art"}
	gradeLevels := []string{"1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th"}

	classrooms := []models.Classroom{}
	for i := 0; i < 30; i++ {
		classroom := models.Classroom{
			Name:       fmt.Sprintf("Classroom %s-%s-%d", subjects[i%len(subjects)], gradeLevels[i%len(gradeLevels)], i+1),
			SchoolID:   schools[i%len(schools)].ID,
			TeacherID:  teachers[i%len(teachers)].ID,
			GradeLevel: gradeLevels[i%len(gradeLevels)],
			Subject:    subjects[i%len(subjects)],
			Capacity:   25 + rand.Intn(10),
		}
		classrooms = append(classrooms, classroom)
	}

	for i := range classrooms {
		if err := db.Create(&classrooms[i]).Error; err != nil {
			return err
		}
	}

	// Create sample students
	students := []models.User{}
	for i := 0; i < 900; i++ { // 30 students per classroom on average
		student := models.User{
			Email:     fmt.Sprintf("student%d@school.edu", i+1),
			Username:  fmt.Sprintf("student%d", i+1),
			FirstName: fmt.Sprintf("Student%d", i+1),
			LastName:  "Johnson",
			Role:      "student",
			SchoolID:  schools[i%len(schools)].ID,
		}
		students = append(students, student)
	}

	for i := range students {
		if err := db.Create(&students[i]).Error; err != nil {
			return err
		}
	}

	// Create enrollments
	for i, student := range students {
		classroomIndex := i / 30 // 30 students per classroom
		if classroomIndex < len(classrooms) {
			enrollment := models.Enrollment{
				ClassroomID: classrooms[classroomIndex].ID,
				UserID:      student.ID,
				EnrolledAt:  time.Now().AddDate(0, -2, 0), // Enrolled 2 months ago
				Status:      "active",
			}
			if err := db.Create(&enrollment).Error; err != nil {
				return err
			}
		}
	}

	// Create sample quizzes
	for _, classroom := range classrooms {
		for j := 0; j < 5; j++ { // 5 quizzes per classroom
			quiz := models.Quiz{
				Title:            fmt.Sprintf("%s Quiz %d", classroom.Subject, j+1),
				ClassroomID:      classroom.ID,
				TeacherID:        classroom.TeacherID,
				QuestionCount:    5,
				TotalPoints:      100.0,
				TimeLimitMinutes: &[]int{30}[0],
				Status:           "published",
				PublishedAt:      &[]time.Time{time.Now().AddDate(0, 0, -rand.Intn(30))}[0],
			}

			if err := db.Create(&quiz).Error; err != nil {
				return err
			}

			// Create sample questions for each quiz
			for k := 0; k < 5; k++ {
				question := models.QuizQuestion{
					QuizID:        quiz.ID,
					QuestionText:  fmt.Sprintf("Sample question %d for %s", k+1, quiz.Title),
					QuestionType:  "multiple_choice",
					Options: models.JSONB{
						"A": "Option A",
						"B": "Option B",
						"C": "Option C",
						"D": "Option D",
					},
					CorrectAnswer: []string{"A", "B", "C", "D"}[rand.Intn(4)],
					Points:        20.0,
					OrderIndex:    k + 1,
				}

				if err := db.Create(&question).Error; err != nil {
					return err
				}

				// Create sample responses
				var enrolledStudents []models.User
				db.Table("users").
					Joins("JOIN enrollments ON users.id = enrollments.user_id").
					Where("enrollments.classroom_id = ? AND users.role = 'student'", classroom.ID).
					Find(&enrolledStudents)

				for _, student := range enrolledStudents {
					// 80% chance student answered the question
					if rand.Float64() < 0.8 {
						isCorrect := rand.Float64() < 0.7 // 70% chance of correct answer
						answer := question.CorrectAnswer
						if !isCorrect {
							// Give wrong answer
							wrongAnswers := []string{"A", "B", "C", "D"}
							for _, wa := range wrongAnswers {
								if wa != question.CorrectAnswer {
									answer = wa
									break
								}
							}
						}

						pointsEarned := 0.0
						if isCorrect {
							pointsEarned = question.Points
						}

						timeTaken := 30 + rand.Intn(120) // 30-150 seconds
						submittedAt := time.Now().AddDate(0, 0, -rand.Intn(15))

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

						if err := db.Create(&response).Error; err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// Create sample sessions and events
	for _, student := range students[:100] { // Create sessions for first 100 students
		for i := 0; i < 10; i++ { // 10 sessions per student
			startTime := time.Now().AddDate(0, 0, -rand.Intn(30))
			duration := 300 + rand.Intn(1800) // 5-35 minutes
			endTime := startTime.Add(time.Duration(duration) * time.Second)

			session := models.Session{
				UserID:          student.ID,
				Application:     []string{"whiteboard", "notebook"}[rand.Intn(2)],
				StartTime:       startTime,
				EndTime:         &endTime,
				DurationSeconds: &duration,
				DeviceType:      []string{"tablet", "laptop", "phone"}[rand.Intn(3)],
				AppVersion:      "2.1.0",
			}

			if err := db.Create(&session).Error; err != nil {
				return err
			}

			// Create sample events for this session
			for j := 0; j < rand.Intn(20)+5; j++ { // 5-25 events per session
				eventTime := startTime.Add(time.Duration(rand.Intn(duration)) * time.Second)

				event := models.Event{
					EventType:   []string{"page_view", "quiz_answer_submitted", "content_created", "note_taken"}[rand.Intn(4)],
					UserID:      student.ID,
					SessionID:   session.ID,
					Timestamp:   eventTime,
					Application: session.Application,
					Payload: models.JSONB{
						"action":     "sample_action",
						"duration":   rand.Intn(60),
						"page_id":    fmt.Sprintf("page_%d", rand.Intn(100)),
					},
					Metadata: models.JSONB{
						"user_agent": "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)",
						"ip_address": "192.168.1.100",
					},
				}

				if err := db.Create(&event).Error; err != nil {
					return err
				}
			}
		}
	}

	log.Printf("Seeded %d schools, %d teachers, %d classrooms, %d students", len(schools), len(teachers), len(classrooms), len(students))
	return nil
}