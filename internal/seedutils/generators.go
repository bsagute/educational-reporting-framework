package seedutils

import (
	"fmt"
	"math/rand"
	"time"

	"reporting-framework/internal/models"

	"gorm.io/gorm"
)

// SchoolGenerator handles the creation of sample school data
type SchoolGenerator struct {
	db     *gorm.DB
	logger Logger
}

// NewSchoolGenerator creates a new school generator instance
func NewSchoolGenerator(db *gorm.DB, logger Logger) *SchoolGenerator {
	return &SchoolGenerator{
		db:     db,
		logger: logger,
	}
}

// GenerateSchools creates a predefined set of sample schools
// Returns the created schools and any error encountered
func (sg *SchoolGenerator) GenerateSchools() ([]models.School, error) {
	sg.logger.Info("Starting school generation process")

	// Define realistic school data based on common naming patterns
	schoolData := []struct {
		name     string
		district string
		region   string
		timezone string
	}{
		{"Lincoln Elementary School", "Downtown District", "North", "America/New_York"},
		{"Washington Middle School", "Suburban District", "South", "America/New_York"},
		{"Roosevelt High School", "Metro District", "East", "America/New_York"},
		{"Jefferson Academy", "Rural District", "West", "America/Chicago"},
		{"Madison Preparatory", "Central District", "Central", "America/Denver"},
	}

	schools := make([]models.School, 0, len(schoolData))

	for _, data := range schoolData {
		school := models.School{
			Name:     data.name,
			District: data.district,
			Region:   data.region,
			Timezone: data.timezone,
		}

		if err := sg.db.Create(&school).Error; err != nil {
			sg.logger.Error("Failed to create school", "school", data.name, "error", err)
			return nil, fmt.Errorf("failed to create school %s: %w", data.name, err)
		}

		schools = append(schools, school)
		sg.logger.Debug("Created school", "name", data.name, "id", school.ID)
	}

	sg.logger.Info("Successfully generated schools", "count", len(schools))
	return schools, nil
}

// UserGenerator handles creation of teachers and students
type UserGenerator struct {
	db     *gorm.DB
	logger Logger
	rand   *rand.Rand
}

// NewUserGenerator creates a new user generator with its own random source
func NewUserGenerator(db *gorm.DB, logger Logger) *UserGenerator {
	// Create a new random source to avoid conflicts with other generators
	source := rand.NewSource(time.Now().UnixNano())
	return &UserGenerator{
		db:     db,
		logger: logger,
		rand:   rand.New(source),
	}
}

// GenerateTeachers creates sample teacher accounts
func (ug *UserGenerator) GenerateTeachers(schools []models.School, count int) ([]models.User, error) {
	ug.logger.Info("Starting teacher generation", "count", count)

	if len(schools) == 0 {
		return nil, fmt.Errorf("cannot generate teachers: no schools provided")
	}

	teachers := make([]models.User, 0, count)

	// Common teacher names for realistic data
	firstNames := []string{"Sarah", "Michael", "Jennifer", "David", "Lisa", "Robert", "Emily", "James", "Amy", "John"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}

	for i := 0; i < count; i++ {
		firstName := firstNames[ug.rand.Intn(len(firstNames))]
		lastName := lastNames[ug.rand.Intn(len(lastNames))]

		teacher := models.User{
			Email:     fmt.Sprintf("%s.%s.%d@school.edu", firstName, lastName, i+1),
			Username:  fmt.Sprintf("%s%s%d", firstName, lastName, i+1),
			FirstName: firstName,
			LastName:  lastName,
			Role:      "teacher",
			SchoolID:  schools[i%len(schools)].ID,
		}

		if err := ug.db.Create(&teacher).Error; err != nil {
			ug.logger.Error("Failed to create teacher", "name", teacher.FirstName+" "+teacher.LastName, "error", err)
			return nil, fmt.Errorf("failed to create teacher %s %s: %w", teacher.FirstName, teacher.LastName, err)
		}

		teachers = append(teachers, teacher)

		if (i+1)%10 == 0 {
			ug.logger.Debug("Teacher generation progress", "created", i+1, "total", count)
		}
	}

	ug.logger.Info("Successfully generated teachers", "count", len(teachers))
	return teachers, nil
}

// GenerateStudents creates sample student accounts
func (ug *UserGenerator) GenerateStudents(schools []models.School, count int) ([]models.User, error) {
	ug.logger.Info("Starting student generation", "count", count)

	if len(schools) == 0 {
		return nil, fmt.Errorf("cannot generate students: no schools provided")
	}

	students := make([]models.User, 0, count)

	// Diverse student names
	firstNames := []string{"Emma", "Liam", "Olivia", "Noah", "Ava", "Ethan", "Sophia", "Mason", "Isabella", "William",
		"Mia", "James", "Charlotte", "Benjamin", "Amelia", "Lucas", "Harper", "Henry", "Evelyn", "Alexander"}
	lastNames := []string{"Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez",
		"Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee"}

	for i := 0; i < count; i++ {
		firstName := firstNames[ug.rand.Intn(len(firstNames))]
		lastName := lastNames[ug.rand.Intn(len(lastNames))]

		student := models.User{
			Email:     fmt.Sprintf("%s.%s.student.%d@school.edu", firstName, lastName, i+1),
			Username:  fmt.Sprintf("student_%s_%s_%d", firstName, lastName, i+1),
			FirstName: firstName,
			LastName:  lastName,
			Role:      "student",
			SchoolID:  schools[i%len(schools)].ID,
		}

		if err := ug.db.Create(&student).Error; err != nil {
			ug.logger.Error("Failed to create student", "name", student.FirstName+" "+student.LastName, "error", err)
			return nil, fmt.Errorf("failed to create student %s %s: %w", student.FirstName, student.LastName, err)
		}

		students = append(students, student)

		// Log progress every 100 students to avoid spam
		if (i+1)%100 == 0 {
			ug.logger.Debug("Student generation progress", "created", i+1, "total", count)
		}
	}

	ug.logger.Info("Successfully generated students", "count", len(students))
	return students, nil
}

// ClassroomGenerator handles classroom creation with realistic academic data
type ClassroomGenerator struct {
	db     *gorm.DB
	logger Logger
	rand   *rand.Rand
}

// NewClassroomGenerator creates a new classroom generator
func NewClassroomGenerator(db *gorm.DB, logger Logger) *ClassroomGenerator {
	source := rand.NewSource(time.Now().UnixNano() + 1000) // Offset to avoid collision
	return &ClassroomGenerator{
		db:     db,
		logger: logger,
		rand:   rand.New(source),
	}
}

// AcademicConfig holds configuration for academic data generation
type AcademicConfig struct {
	Subjects     []string
	GradeLevels  []string
	MinCapacity  int
	MaxCapacity  int
}

// GetDefaultAcademicConfig returns a realistic academic configuration
func GetDefaultAcademicConfig() AcademicConfig {
	return AcademicConfig{
		Subjects:    []string{"Mathematics", "Science", "English Language Arts", "Social Studies", "Art", "Music", "Physical Education", "Computer Science"},
		GradeLevels: []string{"Kindergarten", "1st Grade", "2nd Grade", "3rd Grade", "4th Grade", "5th Grade", "6th Grade", "7th Grade", "8th Grade"},
		MinCapacity: 20,
		MaxCapacity: 35,
	}
}

// GenerateClassrooms creates realistic classroom data
func (cg *ClassroomGenerator) GenerateClassrooms(schools []models.School, teachers []models.User, count int, config AcademicConfig) ([]models.Classroom, error) {
	cg.logger.Info("Starting classroom generation", "count", count)

	if len(schools) == 0 {
		return nil, fmt.Errorf("cannot generate classrooms: no schools provided")
	}
	if len(teachers) == 0 {
		return nil, fmt.Errorf("cannot generate classrooms: no teachers provided")
	}

	classrooms := make([]models.Classroom, 0, count)

	for i := 0; i < count; i++ {
		subject := config.Subjects[i%len(config.Subjects)]
		gradeLevel := config.GradeLevels[i%len(config.GradeLevels)]
		school := schools[i%len(schools)]
		teacher := teachers[i%len(teachers)]

		// Generate realistic classroom names
		classroomName := fmt.Sprintf("%s - %s (Room %d)", subject, gradeLevel, 100+i)
		capacity := config.MinCapacity + cg.rand.Intn(config.MaxCapacity-config.MinCapacity+1)

		classroom := models.Classroom{
			Name:       classroomName,
			SchoolID:   school.ID,
			TeacherID:  teacher.ID,
			GradeLevel: gradeLevel,
			Subject:    subject,
			Capacity:   capacity,
		}

		if err := cg.db.Create(&classroom).Error; err != nil {
			cg.logger.Error("Failed to create classroom", "name", classroomName, "error", err)
			return nil, fmt.Errorf("failed to create classroom %s: %w", classroomName, err)
		}

		classrooms = append(classrooms, classroom)

		if (i+1)%10 == 0 {
			cg.logger.Debug("Classroom generation progress", "created", i+1, "total", count)
		}
	}

	cg.logger.Info("Successfully generated classrooms", "count", len(classrooms))
	return classrooms, nil
}

// EnrollmentManager handles student-classroom relationships
type EnrollmentManager struct {
	db     *gorm.DB
	logger Logger
}

// NewEnrollmentManager creates a new enrollment manager
func NewEnrollmentManager(db *gorm.DB, logger Logger) *EnrollmentManager {
	return &EnrollmentManager{
		db:     db,
		logger: logger,
	}
}

// CreateEnrollments assigns students to classrooms with realistic distribution
func (em *EnrollmentManager) CreateEnrollments(students []models.User, classrooms []models.Classroom, studentsPerClassroom int) error {
	em.logger.Info("Starting enrollment creation", "students", len(students), "classrooms", len(classrooms), "perClassroom", studentsPerClassroom)

	if len(students) == 0 || len(classrooms) == 0 {
		return fmt.Errorf("cannot create enrollments: missing students or classrooms")
	}

	totalEnrollments := 0
	studentIndex := 0

	for i, classroom := range classrooms {
		// Ensure we don't exceed available students
		remainingStudents := len(students) - studentIndex
		if remainingStudents <= 0 {
			em.logger.Warn("Ran out of students for enrollment", "classroom", i+1, "total", len(classrooms))
			break
		}

		// Adjust enrollment count if not enough students remain
		enrollmentCount := studentsPerClassroom
		if remainingStudents < studentsPerClassroom {
			enrollmentCount = remainingStudents
		}

		for j := 0; j < enrollmentCount && studentIndex < len(students); j++ {
			student := students[studentIndex]

			// Create enrollment with some historical variation
			enrolledAt := time.Now().AddDate(0, -2, 0) // Enrolled 2 months ago by default

			enrollment := models.Enrollment{
				ClassroomID: classroom.ID,
				UserID:      student.ID,
				EnrolledAt:  enrolledAt,
				Status:      "active",
			}

			if err := em.db.Create(&enrollment).Error; err != nil {
				em.logger.Error("Failed to create enrollment",
					"student", student.FirstName+" "+student.LastName,
					"classroom", classroom.Name,
					"error", err)
				return fmt.Errorf("failed to enroll student %s in classroom %s: %w",
					student.FirstName+" "+student.LastName, classroom.Name, err)
			}

			studentIndex++
			totalEnrollments++
		}

		if (i+1)%10 == 0 {
			em.logger.Debug("Enrollment progress", "classrooms_processed", i+1, "total_enrollments", totalEnrollments)
		}
	}

	em.logger.Info("Successfully created enrollments", "total", totalEnrollments)
	return nil
}