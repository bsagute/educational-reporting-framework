package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"reporting-framework/internal/domain/reporting"
	"reporting-framework/internal/handlers"
	"reporting-framework/internal/seedutils"
)

func main() {
	// Initialize database connection
	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto-migrate schema
	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Check if we should seed data
	if shouldSeedData() {
		if err := seedDatabase(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	}

	// Initialize HTTP server
	router := setupRouter(db)

	// Start server
	port := getPort()
	fmt.Printf("🚀 Educational Reporting Framework Server starting on port %s\n", port)
	fmt.Printf("📊 API Documentation available at http://localhost:%s/docs\n", port)
	fmt.Printf("🔍 Health check: http://localhost:%s/health\n", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initializeDatabase sets up the database connection
func initializeDatabase() (*gorm.DB, error) {
	dsn := getDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Enable logging in development
		Logger: nil, // You can add a logger here
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	fmt.Println("✅ Database connection established")
	return db, nil
}

// autoMigrate runs database migrations
func autoMigrate(db *gorm.DB) error {
	fmt.Println("🔄 Running database migrations...")

	// Auto-migrate all models
	err := db.AutoMigrate(
		&reporting.School{},
		&reporting.Classroom{},
		&reporting.User{},
		&reporting.UserClassroom{},
		&reporting.Session{},
		&reporting.Content{},
		&reporting.Quiz{},
		&reporting.QuizQuestion{},
		&reporting.QuizSession{},
		&reporting.QuizSubmission{},
		&reporting.Event{},
	)

	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	fmt.Println("✅ Database migrations completed")
	return nil
}

// setupRouter initializes the HTTP router and routes
func setupRouter(db *gorm.DB) *gin.Engine {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Add request logging middleware
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "educational-reporting-framework",
			"version":   "1.0.0",
			"timestamp": nil, // time.Now(),
		})
	})

	// API documentation endpoint
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Educational Reporting Framework API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"events": gin.H{
					"POST /api/v1/events": "Ingest batch events",
					"POST /api/v1/sessions/batch": "Ingest session data with events",
				},
				"reports": gin.H{
					"GET /api/v1/reports/student-performance": "Student performance analytics",
					"GET /api/v1/reports/classroom-engagement": "Classroom engagement metrics",
					"GET /api/v1/reports/content-effectiveness": "Content effectiveness analysis",
					"GET /api/v1/reports/school-overview": "School-level overview",
				},
				"analytics": gin.H{
					"GET /api/v1/analytics/real-time/active-sessions": "Real-time active sessions",
					"GET /api/v1/analytics/trends/engagement": "Engagement trends over time",
					"GET /api/v1/analytics/quiz-analytics/:quiz_id": "Detailed quiz analytics",
				},
				"query": gin.H{
					"POST /api/v1/query": "Generic cube.dev style queries",
					"GET /api/v1/query/schema": "Available measures and dimensions",
				},
				"admin": gin.H{
					"POST /api/v1/admin/schools": "Create school",
					"POST /api/v1/admin/classrooms": "Create classroom",
					"POST /api/v1/admin/users": "Create user",
					"POST /api/v1/admin/refresh-metrics": "Refresh aggregated metrics",
				},
			},
		})
	})

	// Initialize reporting handler
	reportingHandler := handlers.NewReportingHandler(db)

	// Register API routes
	api := router.Group("/api")
	reportingHandler.RegisterRoutes(api)

	// Add schema endpoint for generic queries
	api.GET("/v1/query/schema", func(c *gin.Context) {
		queryBuilder := handlers.NewGenericQueryBuilder(db)
		schema := queryBuilder.GetAvailableMetrics()
		c.JSON(200, schema)
	})

	fmt.Println("✅ HTTP routes registered")
	return router
}

// seedDatabase populates the database with test data
func seedDatabase(db *gorm.DB) error {
	fmt.Println("🌱 Seeding database with test data...")

	seedManager := seedutils.NewSeedManager(db)

	// Check if data already exists
	var schoolCount int64
	db.Model(&reporting.School{}).Count(&schoolCount)

	if schoolCount > 0 {
		fmt.Printf("📊 Database already contains %d schools, skipping seed\n", schoolCount)
		return nil
	}

	return seedManager.SeedAllData()
}

// Helper functions

func getDatabaseDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "reporting_db")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)
}

func getPort() string {
	return getEnv("PORT", "8080")
}

func shouldSeedData() bool {
	return getEnv("SEED_DATA", "true") == "true"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}