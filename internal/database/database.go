package database

import (
	"reporting-framework/internal/models"
	"reporting-framework/internal/seedmigrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate runs database migrations and ensures seed data exists
func Migrate(db *gorm.DB) error {
	// Run schema migrations first
	err := db.AutoMigrate(
		&models.School{},
		&models.User{},
		&models.Classroom{},
		&models.Enrollment{},
		&models.Session{},
		&models.Event{},
		&models.Quiz{},
		&models.QuizQuestion{},
		&models.QuizResponse{},
		&models.DailyUserStats{},
		&models.ClassroomAnalytics{},
	)
	if err != nil {
		return err
	}

	// Run seed migrations to ensure data exists
	// This is non-destructive and only runs if no data exists
	seedmigrations.AutoSeedOnStartup(db)

	return nil
}