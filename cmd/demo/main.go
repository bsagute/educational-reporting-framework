package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"reporting-framework/internal/domain/reporting"
	"reporting-framework/internal/seedutils"
	"reporting-framework/internal/services"
)

func main() {
	fmt.Println("🎯 Educational Reporting Framework - Demo")
	fmt.Println("==========================================")

	// Initialize database
	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto-migrate
	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Seed test data
	if err := seedTestData(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	// Run report demonstrations
	reportsService := services.NewReportsService(db)

	fmt.Println("\n📊 DEMONSTRATING THREE TYPES OF REPORTS")
	fmt.Println("=========================================")

	// 1. Student Performance Analysis
	if err := demonstrateStudentPerformanceReport(reportsService, db); err != nil {
		log.Printf("Error in student performance demo: %v", err)
	}

	// 2. Classroom Engagement Metrics
	if err := demonstrateClassroomEngagementReport(reportsService, db); err != nil {
		log.Printf("Error in classroom engagement demo: %v", err)
	}

	// 3. Content Effectiveness Evaluation
	if err := demonstrateContentEffectivenessReport(reportsService, db); err != nil {
		log.Printf("Error in content effectiveness demo: %v", err)
	}

	// 4. Demonstrate Cube.dev style queries
	if err := demonstrateGenericQueries(db); err != nil {
		log.Printf("Error in generic queries demo: %v", err)
	}

	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Println("📝 Check the generated reports above and the API endpoints at http://localhost:8080")
}

func demonstrateStudentPerformanceReport(reportsService *services.ReportsService, db *gorm.DB) error {
	fmt.Println("\n1️⃣  STUDENT PERFORMANCE ANALYSIS")
	fmt.Println("==================================")

	// Get a random student
	var student reporting.User
	if err := db.Where("role = 'student'").First(&student).Error; err != nil {
		return fmt.Errorf("no students found: %w", err)
	}

	// Generate report for last 30 days
	dateFrom := time.Now().AddDate(0, 0, -30)
	dateTo := time.Now()

	report, err := reportsService.GenerateStudentPerformanceReport(
		student.ID, nil, dateFrom, dateTo)
	if err != nil {
		return fmt.Errorf("failed to generate student report: %w", err)
	}

	fmt.Printf("📚 Student: %s\n", report.StudentName)
	fmt.Printf("📅 Period: %s to %s (%d days)\n",
		report.Period.From.Format("2006-01-02"),
		report.Period.To.Format("2006-01-02"),
		report.Period.Days)

	fmt.Printf("\n📈 PERFORMANCE METRICS:\n")
	fmt.Printf("  • Average Quiz Score: %.1f%%\n", report.OverallStats.AvgQuizScore)
	fmt.Printf("  • Quiz Completion Rate: %.1f%%\n", report.OverallStats.CompletionRate)
	fmt.Printf("  • Engagement Score: %.1f%%\n", report.OverallStats.EngagementScore)
	fmt.Printf("  • Performance Trend: %s\n", report.OverallStats.PerformanceTrend)
	fmt.Printf("  • Active Days: %d/%d\n", report.OverallStats.ActiveDays, report.Period.Days)
	fmt.Printf("  • Avg Daily Minutes: %.1f\n", report.OverallStats.AvgDailyMinutes)

	fmt.Printf("\n💡 RECOMMENDATIONS:\n")
	for i, rec := range report.Recommendations {
		fmt.Printf("  %d. %s\n", i+1, rec)
	}

	// Save report to file
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	filename := fmt.Sprintf("student_performance_report_%s.json", student.ID.String()[:8])
	os.WriteFile(filename, reportJSON, 0644)
	fmt.Printf("\n💾 Full report saved to: %s\n", filename)

	return nil
}

func demonstrateClassroomEngagementReport(reportsService *services.ReportsService, db *gorm.DB) error {
	fmt.Println("\n2️⃣  CLASSROOM ENGAGEMENT METRICS")
	fmt.Println("==================================")

	// Get a random classroom
	var classroom reporting.Classroom
	if err := db.Preload("School").Preload("Teacher").First(&classroom).Error; err != nil {
		return fmt.Errorf("no classrooms found: %w", err)
	}

	// Generate report for last 30 days
	dateFrom := time.Now().AddDate(0, 0, -30)
	dateTo := time.Now()

	report, err := reportsService.GenerateClassroomEngagementReport(
		classroom.ID, dateFrom, dateTo)
	if err != nil {
		return fmt.Errorf("failed to generate classroom report: %w", err)
	}

	fmt.Printf("🏫 Classroom: %s\n", report.ClassroomName)
	fmt.Printf("👨‍🏫 Teacher: %s\n", report.TeacherName)
	fmt.Printf("🏢 School: %s\n", report.SchoolName)
	fmt.Printf("📅 Period: %s to %s\n",
		report.Period.From.Format("2006-01-02"),
		report.Period.To.Format("2006-01-02"))

	metrics := report.EngagementMetrics
	fmt.Printf("\n📊 ENGAGEMENT METRICS:\n")
	fmt.Printf("  • Total Students: %d\n", metrics.TotalStudents)
	fmt.Printf("  • Active Students: %d\n", metrics.ActiveStudents)
	fmt.Printf("  • Participation Rate: %.1f%%\n", metrics.ParticipationRate)
	fmt.Printf("  • Avg Session Duration: %.1f minutes\n", metrics.AvgSessionDuration)
	fmt.Printf("  • Quiz Completion Rate: %.1f%%\n", metrics.AvgQuizCompletionRate)
	fmt.Printf("  • Average Class Score: %.1f%%\n", metrics.AvgClassScore)
	fmt.Printf("  • Overall Engagement: %.1f%%\n", metrics.OverallEngagementScore)

	fmt.Printf("\n📱 PLATFORM USAGE:\n")
	fmt.Printf("  • Whiteboard Usage: %d minutes\n", metrics.WhiteboardUsageMinutes)
	fmt.Printf("  • Notebook Usage: %d minutes\n", metrics.NotebookUsageMinutes)
	fmt.Printf("  • Collaboration Events: %d\n", metrics.CollaborationEvents)
	fmt.Printf("  • Sync Events: %d\n", metrics.SyncEventsCount)

	fmt.Printf("\n🎯 INSIGHTS:\n")
	for i, insight := range report.Insights {
		fmt.Printf("  %d. %s\n", i+1, insight)
	}

	// Save report to file
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	filename := fmt.Sprintf("classroom_engagement_report_%s.json", classroom.ID.String()[:8])
	os.WriteFile(filename, reportJSON, 0644)
	fmt.Printf("\n💾 Full report saved to: %s\n", filename)

	return nil
}

func demonstrateContentEffectivenessReport(reportsService *services.ReportsService, db *gorm.DB) error {
	fmt.Println("\n3️⃣  CONTENT EFFECTIVENESS EVALUATION")
	fmt.Println("======================================")

	// Get a random school
	var school reporting.School
	if err := db.First(&school).Error; err != nil {
		return fmt.Errorf("no schools found: %w", err)
	}

	// Generate report for last 30 days
	dateFrom := time.Now().AddDate(0, 0, -30)
	dateTo := time.Now()

	report, err := reportsService.GenerateContentEffectivenessReport(
		&school.ID, nil, "", dateFrom, dateTo)
	if err != nil {
		return fmt.Errorf("failed to generate content report: %w", err)
	}

	fmt.Printf("🏢 School: %s\n", school.Name)
	fmt.Printf("📅 Analysis Period: %s to %s\n",
		report.Period.From.Format("2006-01-02"),
		report.Period.To.Format("2006-01-02"))

	analytics := report.ContentAnalytics
	fmt.Printf("\n📈 CONTENT ANALYTICS:\n")
	fmt.Printf("  • Total Content Items: %d\n", analytics.TotalContent)
	fmt.Printf("  • Total Views: %d\n", analytics.TotalViews)
	fmt.Printf("  • Unique Viewers: %d\n", analytics.UniqueViewers)
	fmt.Printf("  • Avg View Duration: %.1f seconds\n", analytics.AvgViewDuration)
	fmt.Printf("  • Avg Engagement Score: %.1f%%\n", analytics.AvgEngagementScore)
	fmt.Printf("  • Share Rate: %.1f%%\n", analytics.ShareRate)
	fmt.Printf("  • Interaction Rate: %.1f%%\n", analytics.InteractionRate)

	fmt.Printf("\n📊 CONTENT TYPE PERFORMANCE:\n")
	for i, typeMetric := range report.ContentTypeBreakdown {
		fmt.Printf("  %d. %s: %.1f avg views, %.1f%% effectiveness\n",
			i+1, typeMetric.ContentType, typeMetric.AvgViews, typeMetric.AvgEffectiveness)
	}

	fmt.Printf("\n💡 RECOMMENDATIONS:\n")
	for i, rec := range report.Recommendations {
		fmt.Printf("  %d. [%s] %s\n", i+1, rec.Priority, rec.Description)
	}

	// Save report to file
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	filename := fmt.Sprintf("content_effectiveness_report_%s.json", school.ID.String()[:8])
	os.WriteFile(filename, reportJSON, 0644)
	fmt.Printf("\n💾 Full report saved to: %s\n", filename)

	return nil
}

func demonstrateGenericQueries(db *gorm.DB) error {
	fmt.Println("\n4️⃣  CUBE.DEV STYLE GENERIC QUERIES")
	fmt.Println("===================================")

	// Example 1: Event counts by application type
	fmt.Println("\n📊 Query 1: Event counts by application type (last 7 days)")
	query1 := `
		SELECT
			COALESCE(application, 'unknown') as application,
			COUNT(*) as event_count,
			COUNT(DISTINCT user_id) as unique_users
		FROM events
		WHERE timestamp >= NOW() - INTERVAL '7 days'
		GROUP BY application
		ORDER BY event_count DESC
	`

	var result1 []struct {
		Application string `json:"application"`
		EventCount  int    `json:"event_count"`
		UniqueUsers int    `json:"unique_users"`
	}

	if err := db.Raw(query1).Scan(&result1).Error; err != nil {
		return fmt.Errorf("query 1 failed: %w", err)
	}

	fmt.Printf("Results:\n")
	for _, row := range result1 {
		fmt.Printf("  • %s: %d events, %d unique users\n",
			row.Application, row.EventCount, row.UniqueUsers)
	}

	// Example 2: Daily engagement trends
	fmt.Println("\n📈 Query 2: Daily engagement trends (last 7 days)")
	query2 := `
		SELECT
			DATE(timestamp) as date,
			COUNT(*) as daily_events,
			COUNT(DISTINCT user_id) as active_users,
			COUNT(DISTINCT session_id) as sessions
		FROM events
		WHERE timestamp >= NOW() - INTERVAL '7 days'
		GROUP BY DATE(timestamp)
		ORDER BY date DESC
	`

	var result2 []struct {
		Date        time.Time `json:"date"`
		DailyEvents int       `json:"daily_events"`
		ActiveUsers int       `json:"active_users"`
		Sessions    int       `json:"sessions"`
	}

	if err := db.Raw(query2).Scan(&result2).Error; err != nil {
		return fmt.Errorf("query 2 failed: %w", err)
	}

	fmt.Printf("Results:\n")
	for _, row := range result2 {
		fmt.Printf("  • %s: %d events, %d users, %d sessions\n",
			row.Date.Format("2006-01-02"), row.DailyEvents, row.ActiveUsers, row.Sessions)
	}

	// Example 3: Quiz performance by classroom
	fmt.Println("\n🎯 Query 3: Quiz performance by classroom")
	query3 := `
		SELECT
			c.name as classroom_name,
			c.subject,
			COUNT(qs.id) as quiz_sessions,
			AVG(qs.percentage_score) as avg_score,
			AVG(qs.time_spent_seconds / 60.0) as avg_time_minutes
		FROM quiz_sessions qs
		JOIN quizzes q ON qs.quiz_id = q.id
		JOIN classrooms c ON q.classroom_id = c.id
		WHERE qs.is_completed = true
		GROUP BY c.id, c.name, c.subject
		ORDER BY avg_score DESC
		LIMIT 5
	`

	var result3 []struct {
		ClassroomName   string   `json:"classroom_name"`
		Subject         *string  `json:"subject"`
		QuizSessions    int      `json:"quiz_sessions"`
		AvgScore        *float64 `json:"avg_score"`
		AvgTimeMinutes  *float64 `json:"avg_time_minutes"`
	}

	if err := db.Raw(query3).Scan(&result3).Error; err != nil {
		return fmt.Errorf("query 3 failed: %w", err)
	}

	fmt.Printf("Results (Top 5 performing classrooms):\n")
	for i, row := range result3 {
		subject := "N/A"
		if row.Subject != nil {
			subject = *row.Subject
		}
		avgScore := 0.0
		if row.AvgScore != nil {
			avgScore = *row.AvgScore
		}
		avgTime := 0.0
		if row.AvgTimeMinutes != nil {
			avgTime = *row.AvgTimeMinutes
		}
		fmt.Printf("  %d. %s (%s): %.1f%% avg score, %.1f min avg time, %d sessions\n",
			i+1, row.ClassroomName, subject, avgScore, avgTime, row.QuizSessions)
	}

	return nil
}

func initializeDatabase() (*gorm.DB, error) {
	dsn := getDatabaseDSN()
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
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
}

func seedTestData(db *gorm.DB) error {
	// Check if data already exists
	var count int64
	db.Model(&reporting.School{}).Count(&count)
	if count > 0 {
		fmt.Printf("📊 Database already contains %d schools, skipping seed\n", count)
		return nil
	}

	fmt.Println("🌱 Seeding test data...")
	seedManager := seedutils.NewSeedManager(db)
	return seedManager.SeedAllData()
}

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}