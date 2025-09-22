package seedmigrations

import (
	"fmt"
	"log"
	"time"

	"reporting-framework/internal/models"
	"reporting-framework/internal/seedutils"

	"gorm.io/gorm"
)

// SeedMigration represents a single data seeding operation
type SeedMigration struct {
	ID          string    `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Description string
	ExecutedAt  time.Time `gorm:"not null"`
	Version     string    `gorm:"not null"`
}

// SeedMigrationManager handles the execution and tracking of seed migrations
type SeedMigrationManager struct {
	db     *gorm.DB
	logger seedutils.Logger
}

// NewSeedMigrationManager creates a new seed migration manager
func NewSeedMigrationManager(db *gorm.DB) *SeedMigrationManager {
	return &SeedMigrationManager{
		db:     db,
		logger: seedutils.NewSimpleLogger(),
	}
}

// MigrationFunc represents a function that performs data seeding
type MigrationFunc func(db *gorm.DB, logger seedutils.Logger) error

// SeedMigrationDefinition defines a migration with its metadata
type SeedMigrationDefinition struct {
	ID          string
	Name        string
	Description string
	Version     string
	Execute     MigrationFunc
}

// InitializeSeedMigrations ensures the seed migration tracking table exists
func (smm *SeedMigrationManager) InitializeSeedMigrations() error {
	// Create the seed_migrations table if it doesn't exist
	err := smm.db.AutoMigrate(&SeedMigration{})
	if err != nil {
		return fmt.Errorf("failed to create seed migrations table: %w", err)
	}

	smm.logger.Debug("Seed migration tracking initialized")
	return nil
}

// RunPendingMigrations executes all migrations that haven't been run yet
func (smm *SeedMigrationManager) RunPendingMigrations() error {
	migrations := smm.getAvailableMigrations()

	for _, migration := range migrations {
		executed, err := smm.isMigrationExecuted(migration.ID)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.ID, err)
		}

		if !executed {
			err = smm.executeMigration(migration)
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
			}
		}
	}

	smm.logger.Info("All seed migrations completed successfully")
	return nil
}

// EnsureBasicDataExists checks if essential data exists and creates it if missing
func (smm *SeedMigrationManager) EnsureBasicDataExists() error {
	// Check if we have any schools - this indicates the system has been seeded
	var schoolCount int64
	err := smm.db.Model(&models.School{}).Count(&schoolCount).Error
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if schoolCount == 0 {
		smm.logger.Info("No existing data found, running initial seed migration")
		return smm.RunPendingMigrations()
	}

	smm.logger.Debug("Existing data detected, skipping automatic seeding", "schools", schoolCount)
	return nil
}

// getAvailableMigrations returns all available seed migrations in order
func (smm *SeedMigrationManager) getAvailableMigrations() []SeedMigrationDefinition {
	return []SeedMigrationDefinition{
		{
			ID:          "001_initial_educational_data",
			Name:        "Initial Educational Data",
			Description: "Creates basic schools, users, classrooms, and sample educational content",
			Version:     "1.0.0",
			Execute:     smm.executeInitialEducationalData,
		},
		{
			ID:          "002_sample_quiz_responses",
			Name:        "Sample Quiz Responses",
			Description: "Generates realistic quiz responses and student engagement data",
			Version:     "1.0.0",
			Execute:     smm.executeSampleQuizResponses,
		},
		{
			ID:          "003_usage_analytics_data",
			Name:        "Usage Analytics Data",
			Description: "Creates session and event data for analytics demonstration",
			Version:     "1.0.0",
			Execute:     smm.executeUsageAnalyticsData,
		},
	}
}

// isMigrationExecuted checks if a migration has already been run
func (smm *SeedMigrationManager) isMigrationExecuted(migrationID string) (bool, error) {
	var count int64
	err := smm.db.Model(&SeedMigration{}).Where("id = ?", migrationID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// executeMigration runs a single migration and records its execution
func (smm *SeedMigrationManager) executeMigration(migration SeedMigrationDefinition) error {
	smm.logger.Info("Executing seed migration", "id", migration.ID, "name", migration.Name)

	// Execute the migration function
	err := migration.Execute(smm.db, smm.logger)
	if err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	// Record the successful execution
	seedMigration := SeedMigration{
		ID:          migration.ID,
		Name:        migration.Name,
		Description: migration.Description,
		ExecutedAt:  time.Now(),
		Version:     migration.Version,
	}

	err = smm.db.Create(&seedMigration).Error
	if err != nil {
		return fmt.Errorf("failed to record migration execution: %w", err)
	}

	smm.logger.Info("Seed migration completed successfully", "id", migration.ID)
	return nil
}

// executeInitialEducationalData creates the foundation of educational data
func (smm *SeedMigrationManager) executeInitialEducationalData(db *gorm.DB, logger seedutils.Logger) error {
	logger.Info("Creating initial educational foundation")

	// Create schools
	schoolGen := seedutils.NewSchoolGenerator(db, logger)
	schools, err := schoolGen.GenerateSchools()
	if err != nil {
		return fmt.Errorf("failed to generate schools: %w", err)
	}

	// Create users (teachers and students)
	userGen := seedutils.NewUserGenerator(db, logger)

	teachers, err := userGen.GenerateTeachers(schools, 25) // 5 teachers per school
	if err != nil {
		return fmt.Errorf("failed to generate teachers: %w", err)
	}

	students, err := userGen.GenerateStudents(schools, 375) // 75 students per school
	if err != nil {
		return fmt.Errorf("failed to generate students: %w", err)
	}

	// Create classrooms
	classGen := seedutils.NewClassroomGenerator(db, logger)
	academicConfig := seedutils.GetDefaultAcademicConfig()

	classrooms, err := classGen.GenerateClassrooms(schools, teachers, 25, academicConfig) // 5 classrooms per school
	if err != nil {
		return fmt.Errorf("failed to generate classrooms: %w", err)
	}

	// Create enrollments
	enrollmentManager := seedutils.NewEnrollmentManager(db, logger)
	err = enrollmentManager.CreateEnrollments(students, classrooms, 15) // 15 students per classroom
	if err != nil {
		return fmt.Errorf("failed to create enrollments: %w", err)
	}

	// Create quizzes
	quizGen := seedutils.NewQuizGenerator(db, logger)
	_, err = quizGen.GenerateQuizzesForClassrooms(classrooms, 6) // 6 quizzes per classroom
	if err != nil {
		return fmt.Errorf("failed to generate quizzes: %w", err)
	}

	logger.Info("Initial educational data migration completed successfully")
	return nil
}

// executeSampleQuizResponses generates realistic student responses
func (smm *SeedMigrationManager) executeSampleQuizResponses(db *gorm.DB, logger seedutils.Logger) error {
	logger.Info("Generating sample quiz responses")

	// Get all quizzes
	var quizzes []models.Quiz
	err := db.Find(&quizzes).Error
	if err != nil {
		return fmt.Errorf("failed to retrieve quizzes: %w", err)
	}

	// Generate responses
	responseGen := seedutils.NewResponseGenerator(db, logger)
	err = responseGen.GenerateResponsesForAllQuizzes(quizzes)
	if err != nil {
		return fmt.Errorf("failed to generate quiz responses: %w", err)
	}

	logger.Info("Sample quiz responses migration completed successfully")
	return nil
}

// executeUsageAnalyticsData creates session and event data
func (smm *SeedMigrationManager) executeUsageAnalyticsData(db *gorm.DB, logger seedutils.Logger) error {
	logger.Info("Generating usage analytics data")

	// Get a subset of students for session generation
	var students []models.User
	err := db.Where("role = ?", "student").Limit(100).Find(&students).Error
	if err != nil {
		return fmt.Errorf("failed to retrieve students: %w", err)
	}

	// Generate sessions
	sessionGen := seedutils.NewSessionGenerator(db, logger)
	sessionPattern := seedutils.GetEducationalSessionPattern()
	sessionPattern.SessionsPerUser = 10 // 10 sessions per student

	sessions, err := sessionGen.GenerateSessionsForUsers(students, sessionPattern)
	if err != nil {
		return fmt.Errorf("failed to generate sessions: %w", err)
	}

	// Generate events
	eventGen := seedutils.NewEventGenerator(db, logger)
	eventPattern := seedutils.GetEducationalEventPattern()
	err = eventGen.GenerateEventsForSessions(sessions, eventPattern)
	if err != nil {
		return fmt.Errorf("failed to generate events: %w", err)
	}

	logger.Info("Usage analytics data migration completed successfully")
	return nil
}

// GetSeedStatus returns information about executed migrations
func (smm *SeedMigrationManager) GetSeedStatus() ([]SeedMigration, error) {
	var migrations []SeedMigration
	err := smm.db.Order("executed_at DESC").Find(&migrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve migration status: %w", err)
	}
	return migrations, nil
}

// AutoSeedOnStartup runs the seed migration system during application startup
// This ensures the system always has data available for development and testing
func AutoSeedOnStartup(db *gorm.DB) {
	manager := NewSeedMigrationManager(db)

	// Initialize the migration tracking system
	err := manager.InitializeSeedMigrations()
	if err != nil {
		log.Printf("Warning: Failed to initialize seed migrations: %v", err)
		return
	}

	// Ensure basic data exists (non-destructive)
	err = manager.EnsureBasicDataExists()
	if err != nil {
		log.Printf("Warning: Failed to ensure basic data exists: %v", err)
		return
	}

	log.Println("Seed migration system completed successfully")
}