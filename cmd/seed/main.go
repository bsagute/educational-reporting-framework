package main

import (
	"os"

	"reporting-framework/internal/config"
	"reporting-framework/internal/database"

	"github.com/joho/godotenv"
)

// DataSeeder orchestrates the overall data seeding process for the reporting framework
type DataSeeder struct{}

// NewDataSeeder creates a new instance of the data seeding orchestrator
func NewDataSeeder() *DataSeeder {
	return &DataSeeder{}
}

// main is the entry point for the database seeding utility
// It coordinates environment setup, database connectivity, and data generation
func main() {
	seeder := NewDataSeeder()

	// Attempt to load environment variables from .env file
	// This is optional - the application can function with environment variables set directly
	if err := godotenv.Load(); err != nil {
		// Environment file not found, using system environment variables
	}

	// Load application configuration from environment
	cfg := config.Load()
	if cfg == nil {
		os.Exit(1)
	}

	// Execute the complete seeding workflow
	if err := seeder.ExecuteSeedingWorkflow(cfg); err != nil {
		os.Exit(1)
	}
}

// ExecuteSeedingWorkflow handles the complete data seeding process
// This method coordinates database setup, migration verification, and data generation
func (ds *DataSeeder) ExecuteSeedingWorkflow(cfg *config.Config) error {
	// Initialize database connection
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		return err
	}

	// Ensure database schema is current before seeding
	// The Migrate function now includes auto-seeding via seedmigrations
	if err := database.Migrate(db); err != nil {
		return err
	}

	return nil
}
