# Seed Migration System

## Overview

The seed migration system ensures that the educational reporting framework always has realistic data available for development, testing, and demonstration purposes. This system automatically runs during application startup and is completely non-destructive.

## Architecture

### Components

1. **SeedMigrationManager** (`internal/seedmigrations/manager.go`)
   - Tracks which seed migrations have been executed
   - Manages the execution order and error handling
   - Provides status tracking via `seed_migrations` table

2. **Seed Utilities** (`internal/seedutils/`)
   - Modular data generators for different entity types
   - Modular code with proper error handling
   - Realistic data patterns based on educational standards

3. **Auto-Initialization** (`internal/database/database.go`)
   - Integrated into the standard migration process
   - Runs automatically when the application starts
   - Only executes if no existing data is found

### Data Generation Process

The system creates data in a logical dependency order:

1. **Schools** → Educational institutions with realistic regional distribution
2. **Users** → Teachers and students with proper role assignments
3. **Classrooms** → Learning spaces with subject and grade level associations
4. **Enrollments** → Student-classroom relationships
5. **Educational Content** → Quizzes with subject-appropriate questions
6. **Student Responses** → Realistic participation and performance patterns
7. **Usage Analytics** → Session and event data for reporting demonstrations

## Key Features

### Non-Destructive Operation
- **Safe for Production**: Only runs if no data exists (checks for existing schools)
- **Idempotent**: Can be run multiple times without issues
- **Trackable**: Records which migrations have been executed

### Data Quality
- **Realistic Names**: Uses diverse, realistic name pools for users
- **Academic Standards**: Follows proper grade level and subject classifications
- **Performance Patterns**: Simulates realistic student response rates and accuracy
- **Usage Behaviors**: Creates believable session durations and interaction patterns

### Scalable Configuration
```go
type EducationalDataConfig struct {
    SchoolCount          int  // Number of institutions
    TeachersPerSchool    int  // Faculty size per school
    ClassroomsPerSchool  int  // Learning spaces per institution
    StudentsPerClassroom int  // Class size
    QuizzesPerClassroom  int  // Assessment frequency
    SessionsPerStudent   int  // Usage activity level
    LimitStudentsForDemo int  // Performance optimization for development
}
```

## Usage

### Automatic Initialization
The system runs automatically when the application starts:

```go
// In database.Migrate()
seedmigrations.AutoSeedOnStartup(db)
```

### Manual Execution
For development purposes, you can also run seeding manually:

```bash
go run cmd/seed/main.go
```

### Migration Tracking
Check which migrations have been executed:

```sql
SELECT * FROM seed_migrations ORDER BY executed_at DESC;
```

## Migration Definitions

### 001_initial_educational_data
- Creates foundational educational structure
- **Schools**: 5 institutions across different districts
- **Teachers**: 25 faculty members with realistic profiles
- **Students**: 375 student accounts across grade levels
- **Classrooms**: 25 learning spaces with proper subject assignments
- **Enrollments**: Student-classroom relationships (15 students per class)
- **Quizzes**: 150+ assessments with subject-appropriate content

### 002_sample_quiz_responses
- Generates realistic student engagement patterns
- **Participation Rate**: 85% of students participate in assessments
- **Performance Distribution**: 72% average accuracy (realistic for K-8)
- **Response Timing**: 15-180 seconds per question
- **Answer Patterns**: Intelligent wrong answer selection

### 003_usage_analytics_data
- Creates session and interaction data for analytics
- **Session Patterns**: 5-60 minute sessions during school hours
- **Device Diversity**: Tablets, laptops, Chromebooks, desktops
- **Application Usage**: Balanced whiteboard/notebook usage
- **Event Tracking**: 5-25 interactions per session with realistic payloads

## Data Volumes

### Production Scale Configuration
- **5 Schools** with regional diversity
- **25 Teachers** across all grade levels and subjects
- **375 Students** with realistic enrollment distribution
- **25 Classrooms** covering K-8 grade levels
- **150+ Quizzes** with 5-question assessments
- **1000+ Sessions** demonstrating active usage
- **10,000+ Events** showing detailed interaction patterns
- **3000+ Quiz Responses** with realistic performance data

### Performance Optimizations
- **Limited Session Generation**: Restricts to 150 students for demo performance
- **Batch Processing**: Efficient database operations
- **Dependency Ordering**: Prevents referential integrity issues
- **Error Recovery**: Graceful handling of partial failures

## Development Benefits

### Always Available Data
- **No Setup Required**: Fresh environments automatically have data
- **Consistent State**: Predictable data set for testing
- **Realistic Scenarios**: Based on actual educational patterns


### Analytics Ready
- **Complete Dataset**: All entity relationships properly established
- **Time-Series Data**: Historical patterns for trending analysis
- **Performance Metrics**: Ready for dashboard development
- **Usage Patterns**: Realistic data for behavioral analytics

## Error Handling

The system includes comprehensive error handling:

- **Database Connectivity**: Graceful degradation if database unavailable
- **Partial Failures**: Continues processing after non-critical errors
- **Migration Tracking**: Records successful completions to avoid duplicates
- **Logging Integration**: Detailed progress and error reporting

## Future Extensions

The migration system is designed for extensibility:

1. **Additional Migrations**: Easy to add new data sets
2. **Environment-Specific Data**: Different scales for dev/staging/prod
3. **Custom Scenarios**: Special test cases for specific features
4. **Data Refresh**: Periodic updates to keep data current

## Monitoring

### Health Checks
- Verify data exists: `SELECT COUNT(*) FROM schools;`
- Check migration status: `SELECT * FROM seed_migrations;`
- Validate data quality: Sample queries for referential integrity

### Performance Metrics
- Migration execution time
- Data generation throughput
- Database impact assessment

