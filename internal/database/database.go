package database

import (
	"reporting-framework/internal/models"

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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
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
}