package seedutils

import (
	"fmt"

	"reporting-framework/internal/models"

	"gorm.io/gorm"
)

// SeedCoordinator manages the overall data seeding process
// It orchestrates the creation of educational data in a logical sequence
type SeedCoordinator struct {
	db     *gorm.DB
	logger Logger
}

// NewSeedCoordinator creates a new coordinator instance
func NewSeedCoordinator(db *gorm.DB, logger Logger) *SeedCoordinator {
	return &SeedCoordinator{
		db:     db,
		logger: logger,
	}
}

// EducationalDataConfig defines the scale and scope of data to be generated
type EducationalDataConfig struct {
	SchoolCount          int
	TeachersPerSchool    int
	ClassroomsPerSchool  int
	StudentsPerClassroom int
	QuizzesPerClassroom  int
	SessionsPerStudent   int
	LimitStudentsForDemo int // Limit session generation to avoid excessive data
}

// GetProductionDataConfig returns configuration suitable for production-scale testing
func GetProductionDataConfig() EducationalDataConfig {
	return EducationalDataConfig{
		SchoolCount:          5,   // Realistic number for initial deployment
		TeachersPerSchool:    10,  // 50 teachers total
		ClassroomsPerSchool:  15,  // 75 classrooms total
		StudentsPerClassroom: 25,  // Realistic class size
		QuizzesPerClassroom:  8,   // 2 months of weekly quizzes
		SessionsPerStudent:   12,  // Represents active usage over time
		LimitStudentsForDemo: 150, // Limit session generation for demo purposes
	}
}

// GenerateEducationalDataset creates a complete educational dataset
// This method follows a dependency-aware sequence to ensure referential integrity
func (sc *SeedCoordinator) GenerateEducationalDataset() error {
	sc.logger.Info("Initializing educational dataset generation")

	config := GetProductionDataConfig()

	// Phase 1: Create foundational institutional data
	schools, err := sc.createSchoolFoundation()
	if err != nil {
		return fmt.Errorf("failed to create school foundation: %w", err)
	}

	// Phase 2: Generate human resources (teachers and students)
	teachers, students, err := sc.createHumanResources(schools, config)
	if err != nil {
		return fmt.Errorf("failed to create human resources: %w", err)
	}

	// Phase 3: Establish learning environments
	classrooms, err := sc.createLearningEnvironments(schools, teachers, config)
	if err != nil {
		return fmt.Errorf("failed to create learning environments: %w", err)
	}

	// Phase 4: Connect students with classrooms
	if err := sc.establishEnrollments(students, classrooms, config); err != nil {
		return fmt.Errorf("failed to establish enrollments: %w", err)
	}

	// Phase 5: Generate educational content and assessments
	quizzes, err := sc.createEducationalContent(classrooms, config)
	if err != nil {
		return fmt.Errorf("failed to create educational content: %w", err)
	}

	// Phase 6: Simulate student responses and interactions
	if err := sc.simulateStudentEngagement(quizzes); err != nil {
		return fmt.Errorf("failed to simulate student engagement: %w", err)
	}

	// Phase 7: Generate usage analytics data
	if err := sc.generateUsageAnalytics(students, config); err != nil {
		return fmt.Errorf("failed to generate usage analytics: %w", err)
	}

	sc.logDatasetSummary(schools, teachers, students, classrooms, quizzes)
	return nil
}

// createSchoolFoundation generates the institutional foundation
func (sc *SeedCoordinator) createSchoolFoundation() ([]models.School, error) {
	sc.logger.Info("Creating institutional foundation")

	generator := NewSchoolGenerator(sc.db, sc.logger)
	schools, err := generator.GenerateSchools()
	if err != nil {
		return nil, fmt.Errorf("school generation failed: %w", err)
	}

	sc.logger.Info("Institutional foundation established", "schools", len(schools))
	return schools, nil
}

// createHumanResources generates teachers and students
func (sc *SeedCoordinator) createHumanResources(schools []models.School, config EducationalDataConfig) ([]models.User, []models.User, error) {
	sc.logger.Info("Creating human resources")

	userGen := NewUserGenerator(sc.db, sc.logger)

	// Generate teaching staff
	totalTeachers := len(schools) * config.TeachersPerSchool
	teachers, err := userGen.GenerateTeachers(schools, totalTeachers)
	if err != nil {
		return nil, nil, fmt.Errorf("teacher generation failed: %w", err)
	}

	// Generate student population
	totalClassrooms := len(schools) * config.ClassroomsPerSchool
	totalStudents := totalClassrooms * config.StudentsPerClassroom
	students, err := userGen.GenerateStudents(schools, totalStudents)
	if err != nil {
		return nil, nil, fmt.Errorf("student generation failed: %w", err)
	}

	sc.logger.Info("Human resources created", "teachers", len(teachers), "students", len(students))
	return teachers, students, nil
}

// createLearningEnvironments establishes classrooms with proper academic structure
func (sc *SeedCoordinator) createLearningEnvironments(schools []models.School, teachers []models.User, config EducationalDataConfig) ([]models.Classroom, error) {
	sc.logger.Info("Creating learning environments")

	classGen := NewClassroomGenerator(sc.db, sc.logger)
	academicConfig := GetDefaultAcademicConfig()

	totalClassrooms := len(schools) * config.ClassroomsPerSchool
	classrooms, err := classGen.GenerateClassrooms(schools, teachers, totalClassrooms, academicConfig)
	if err != nil {
		return nil, fmt.Errorf("classroom generation failed: %w", err)
	}

	sc.logger.Info("Learning environments established", "classrooms", len(classrooms))
	return classrooms, nil
}

// establishEnrollments creates student-classroom relationships
func (sc *SeedCoordinator) establishEnrollments(students []models.User, classrooms []models.Classroom, config EducationalDataConfig) error {
	sc.logger.Info("Establishing student enrollments")

	enrollmentManager := NewEnrollmentManager(sc.db, sc.logger)
	err := enrollmentManager.CreateEnrollments(students, classrooms, config.StudentsPerClassroom)
	if err != nil {
		return fmt.Errorf("enrollment creation failed: %w", err)
	}

	sc.logger.Info("Student enrollments established successfully")
	return nil
}

// createEducationalContent generates quizzes and questions
func (sc *SeedCoordinator) createEducationalContent(classrooms []models.Classroom, config EducationalDataConfig) ([]models.Quiz, error) {
	sc.logger.Info("Creating educational content")

	quizGen := NewQuizGenerator(sc.db, sc.logger)
	quizzes, err := quizGen.GenerateQuizzesForClassrooms(classrooms, config.QuizzesPerClassroom)
	if err != nil {
		return nil, fmt.Errorf("quiz generation failed: %w", err)
	}

	sc.logger.Info("Educational content created", "quizzes", len(quizzes))
	return quizzes, nil
}

// simulateStudentEngagement creates realistic quiz responses
func (sc *SeedCoordinator) simulateStudentEngagement(quizzes []models.Quiz) error {
	sc.logger.Info("Simulating student engagement patterns")

	responseGen := NewResponseGenerator(sc.db, sc.logger)
	err := responseGen.GenerateResponsesForAllQuizzes(quizzes)
	if err != nil {
		return fmt.Errorf("response generation failed: %w", err)
	}

	sc.logger.Info("Student engagement simulation completed")
	return nil
}

// generateUsageAnalytics creates session and event data for analytics
func (sc *SeedCoordinator) generateUsageAnalytics(students []models.User, config EducationalDataConfig) error {
	sc.logger.Info("Generating usage analytics data")

	// Generate realistic session patterns
	sessionGen := NewSessionGenerator(sc.db, sc.logger)
	sessionPattern := GetEducationalSessionPattern()
	sessionPattern.SessionsPerUser = config.SessionsPerStudent

	// Limit session generation to avoid excessive data during development
	limitedStudents := students
	if len(students) > config.LimitStudentsForDemo {
		limitedStudents = students[:config.LimitStudentsForDemo]
		sc.logger.Info("Limiting session generation for performance", "limited_to", config.LimitStudentsForDemo)
	}

	sessions, err := sessionGen.GenerateSessionsForUsers(limitedStudents, sessionPattern)
	if err != nil {
		return fmt.Errorf("session generation failed: %w", err)
	}

	// Generate events within sessions
	eventGen := NewEventGenerator(sc.db, sc.logger)
	eventPattern := GetEducationalEventPattern()
	err = eventGen.GenerateEventsForSessions(sessions, eventPattern)
	if err != nil {
		return fmt.Errorf("event generation failed: %w", err)
	}

	sc.logger.Info("Usage analytics data generation completed", "sessions", len(sessions))
	return nil
}

// logDatasetSummary provides a comprehensive overview of generated data
func (sc *SeedCoordinator) logDatasetSummary(schools []models.School, teachers []models.User, students []models.User, classrooms []models.Classroom, quizzes []models.Quiz) {
	sc.logger.Info("=== Educational Dataset Generation Summary ===")
	sc.logger.Info("Dataset composition", "schools", len(schools))
	sc.logger.Info("Human resources", "teachers", len(teachers), "students", len(students))
	sc.logger.Info("Learning infrastructure", "classrooms", len(classrooms), "quizzes", len(quizzes))

	// Calculate some interesting metrics
	avgStudentsPerSchool := len(students) / len(schools)
	avgClassroomsPerSchool := len(classrooms) / len(schools)
	avgQuizzesPerClassroom := len(quizzes) / len(classrooms)

	sc.logger.Info("Dataset ratios",
		"avg_students_per_school", avgStudentsPerSchool,
		"avg_classrooms_per_school", avgClassroomsPerSchool,
		"avg_quizzes_per_classroom", avgQuizzesPerClassroom)

	sc.logger.Info("Educational dataset ready for analytics and reporting")
}