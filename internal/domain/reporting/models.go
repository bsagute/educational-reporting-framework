package reporting

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONB represents a PostgreSQL JSONB column
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// School represents an educational institution
type School struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name         string     `json:"name" gorm:"not null"`
	District     *string    `json:"district"`
	Region       *string    `json:"region"`
	ContactEmail *string    `json:"contact_email"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Classrooms []Classroom `json:"classrooms,omitempty" gorm:"foreignKey:SchoolID"`
	Users      []User      `json:"users,omitempty" gorm:"foreignKey:SchoolID"`
}

// Classroom represents a class within a school
type Classroom struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SchoolID    uuid.UUID  `json:"school_id" gorm:"not null"`
	Name        string     `json:"name" gorm:"not null"`
	GradeLevel  *int       `json:"grade_level"`
	Subject     *string    `json:"subject"`
	TeacherID   *uuid.UUID `json:"teacher_id"`
	MaxStudents int        `json:"max_students" gorm:"default:30"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	School   School             `json:"school,omitempty" gorm:"foreignKey:SchoolID"`
	Teacher  *User              `json:"teacher,omitempty" gorm:"foreignKey:TeacherID"`
	Users    []User             `json:"users,omitempty" gorm:"many2many:user_classrooms"`
	Sessions []Session          `json:"sessions,omitempty" gorm:"foreignKey:ClassroomID"`
	Quizzes  []Quiz             `json:"quizzes,omitempty" gorm:"foreignKey:ClassroomID"`
	Content  []Content          `json:"content,omitempty" gorm:"foreignKey:ClassroomID"`
}

// User represents a user in the system (student, teacher, or admin)
type User struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SchoolID   uuid.UUID  `json:"school_id" gorm:"not null"`
	Username   string     `json:"username" gorm:"unique;not null"`
	Email      *string    `json:"email"`
	Role       string     `json:"role" gorm:"not null"` // teacher, student, admin
	FirstName  *string    `json:"first_name"`
	LastName   *string    `json:"last_name"`
	LastActive *time.Time `json:"last_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	School       School             `json:"school,omitempty" gorm:"foreignKey:SchoolID"`
	Classrooms   []Classroom        `json:"classrooms,omitempty" gorm:"many2many:user_classrooms"`
	Sessions     []Session          `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	Events       []Event            `json:"events,omitempty" gorm:"foreignKey:UserID"`
	CreatedQuizzes []Quiz           `json:"created_quizzes,omitempty" gorm:"foreignKey:CreatorID"`
	QuizSessions []QuizSession      `json:"quiz_sessions,omitempty" gorm:"foreignKey:StudentID"`
	Content      []Content          `json:"content,omitempty" gorm:"foreignKey:CreatorID"`
}

// UserClassroom represents the many-to-many relationship between users and classrooms
type UserClassroom struct {
	UserID      uuid.UUID `json:"user_id" gorm:"primaryKey"`
	ClassroomID uuid.UUID `json:"classroom_id" gorm:"primaryKey"`
	Role        string    `json:"role" gorm:"not null"` // teacher, student
	EnrolledAt  time.Time `json:"enrolled_at" gorm:"default:CURRENT_TIMESTAMP"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`

	// Relationships
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Classroom Classroom `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
}

// Session represents a user session in either application
type Session struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"not null"`
	ClassroomID     *uuid.UUID `json:"classroom_id"`
	Application     string     `json:"application" gorm:"not null"` // whiteboard, notebook
	StartTime       time.Time  `json:"start_time" gorm:"not null"`
	EndTime         *time.Time `json:"end_time"`
	DurationSeconds *int       `json:"duration_seconds"`
	DeviceInfo      JSONB      `json:"device_info"`
	IPAddress       *string    `json:"ip_address" gorm:"type:inet"`
	CreatedAt       time.Time  `json:"created_at"`

	// Relationships
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Classroom *Classroom `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	Events    []Event    `json:"events,omitempty" gorm:"foreignKey:SessionID"`
}

// Content represents educational content created in the system
type Content struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatorID       uuid.UUID  `json:"creator_id" gorm:"not null"`
	ClassroomID     *uuid.UUID `json:"classroom_id"`
	Title           *string    `json:"title"`
	ContentType     string     `json:"content_type" gorm:"not null"` // note, drawing, document, quiz, whiteboard_session
	ContentData     JSONB      `json:"content_data"`
	FileSizeBytes   int64      `json:"file_size_bytes" gorm:"default:0"`
	IsShared        bool       `json:"is_shared" gorm:"default:false"`
	SharePermissions JSONB     `json:"share_permissions"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Creator   User       `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Classroom *Classroom `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
}

// Quiz represents a quiz created by a teacher
type Quiz struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatorID     uuid.UUID  `json:"creator_id" gorm:"not null"`
	ClassroomID   uuid.UUID  `json:"classroom_id" gorm:"not null"`
	Title         string     `json:"title" gorm:"not null"`
	Description   *string    `json:"description"`
	TotalQuestions int       `json:"total_questions" gorm:"default:0"`
	TimeLimitMinutes *int    `json:"time_limit_minutes"`
	MaxAttempts   int        `json:"max_attempts" gorm:"default:1"`
	IsActive      bool       `json:"is_active" gorm:"default:false"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Creator   User          `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Classroom Classroom     `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	Questions []QuizQuestion `json:"questions,omitempty" gorm:"foreignKey:QuizID"`
	Sessions  []QuizSession `json:"sessions,omitempty" gorm:"foreignKey:QuizID"`
}

// QuizQuestion represents a question within a quiz
type QuizQuestion struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	QuizID       uuid.UUID  `json:"quiz_id" gorm:"not null"`
	QuestionText string     `json:"question_text" gorm:"not null"`
	QuestionType string     `json:"question_type" gorm:"not null"` // multiple_choice, true_false, short_answer, essay
	Options      JSONB      `json:"options"`
	CorrectAnswer *string   `json:"correct_answer"`
	Points       int        `json:"points" gorm:"default:1"`
	OrderIndex   int        `json:"order_index" gorm:"not null"`
	Explanation  *string    `json:"explanation"`
	CreatedAt    time.Time  `json:"created_at"`

	// Relationships
	Quiz        Quiz             `json:"quiz,omitempty" gorm:"foreignKey:QuizID"`
	Submissions []QuizSubmission `json:"submissions,omitempty" gorm:"foreignKey:QuestionID"`
}

// QuizSession represents a student's attempt at a quiz
type QuizSession struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	QuizID            uuid.UUID  `json:"quiz_id" gorm:"not null"`
	StudentID         uuid.UUID  `json:"student_id" gorm:"not null"`
	StartedAt         time.Time  `json:"started_at" gorm:"default:CURRENT_TIMESTAMP"`
	CompletedAt       *time.Time `json:"completed_at"`
	TotalScore        int        `json:"total_score" gorm:"default:0"`
	MaxPossibleScore  int        `json:"max_possible_score" gorm:"default:0"`
	PercentageScore   *float64   `json:"percentage_score"`
	TimeSpentSeconds  *int       `json:"time_spent_seconds"`
	AttemptNumber     int        `json:"attempt_number" gorm:"default:1"`
	IsCompleted       bool       `json:"is_completed" gorm:"default:false"`

	// Relationships
	Quiz        Quiz             `json:"quiz,omitempty" gorm:"foreignKey:QuizID"`
	Student     User             `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	Submissions []QuizSubmission `json:"submissions,omitempty" gorm:"foreignKey:QuizID,StudentID;references:QuizID,StudentID"`
}

// QuizSubmission represents a student's answer to a quiz question
type QuizSubmission struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	QuizID          uuid.UUID  `json:"quiz_id" gorm:"not null"`
	StudentID       uuid.UUID  `json:"student_id" gorm:"not null"`
	QuestionID      uuid.UUID  `json:"question_id" gorm:"not null"`
	SubmittedAnswer *string    `json:"submitted_answer"`
	IsCorrect       *bool      `json:"is_correct"`
	PointsEarned    int        `json:"points_earned" gorm:"default:0"`
	TimeSpentSeconds *int      `json:"time_spent_seconds"`
	AttemptNumber   int        `json:"attempt_number" gorm:"default:1"`
	SubmittedAt     time.Time  `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Quiz     Quiz         `json:"quiz,omitempty" gorm:"foreignKey:QuizID"`
	Student  User         `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	Question QuizQuestion `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
}

// Event represents a user interaction event
type Event struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	EventType   string     `json:"event_type" gorm:"not null"`
	UserID      *uuid.UUID `json:"user_id"`
	SessionID   *uuid.UUID `json:"session_id"`
	ClassroomID *uuid.UUID `json:"classroom_id"`
	SchoolID    *uuid.UUID `json:"school_id"`
	Application *string    `json:"application"` // whiteboard, notebook
	Timestamp   time.Time  `json:"timestamp" gorm:"not null"`
	Metadata    JSONB      `json:"metadata"`
	DeviceInfo  JSONB      `json:"device_info"`
	CreatedAt   time.Time  `json:"created_at"`

	// Relationships
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Session   *Session   `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	Classroom *Classroom `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	School    *School    `json:"school,omitempty" gorm:"foreignKey:SchoolID"`
}

// EventRequest represents the API request structure for event ingestion
type EventRequest struct {
	Events []EventData `json:"events" validate:"required,min=1,max=100"`
}

// EventData represents individual event data for ingestion
type EventData struct {
	EventType   string                 `json:"event_type" validate:"required"`
	Timestamp   time.Time              `json:"timestamp" validate:"required"`
	UserID      *uuid.UUID             `json:"user_id"`
	SessionID   *uuid.UUID             `json:"session_id"`
	ClassroomID *uuid.UUID             `json:"classroom_id"`
	Application *string                `json:"application"`
	Metadata    map[string]interface{} `json:"metadata"`
	DeviceInfo  map[string]interface{} `json:"device_info"`
}

// EventResponse represents the API response for event ingestion
type EventResponse struct {
	Success       bool   `json:"success"`
	ProcessedCount int   `json:"processed_count"`
	Message       string `json:"message"`
	EventIDs      []uuid.UUID `json:"event_ids,omitempty"`
}

// TableName methods for GORM
func (School) TableName() string        { return "schools" }
func (Classroom) TableName() string     { return "classrooms" }
func (User) TableName() string          { return "users" }
func (UserClassroom) TableName() string { return "user_classrooms" }
func (Session) TableName() string       { return "sessions" }
func (Content) TableName() string       { return "content" }
func (Quiz) TableName() string          { return "quizzes" }
func (QuizQuestion) TableName() string  { return "quiz_questions" }
func (QuizSession) TableName() string   { return "quiz_sessions" }
func (QuizSubmission) TableName() string { return "quiz_submissions" }
func (Event) TableName() string         { return "events" }