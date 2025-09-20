package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom JSONB type for PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// School represents an educational institution
type School struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	District  string    `json:"district"`
	Region    string    `json:"region"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *School) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}

// User represents teachers, students, and administrators
type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email      string    `gorm:"unique;not null" json:"email"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Role       string    `gorm:"type:varchar(20);not null" json:"role"` // teacher, student, admin
	SchoolID   uuid.UUID `gorm:"type:uuid" json:"school_id"`
	School     School    `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastActive *time.Time `json:"last_active"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

// Classroom represents learning spaces
type Classroom struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name       string    `gorm:"not null" json:"name"`
	SchoolID   uuid.UUID `gorm:"type:uuid" json:"school_id"`
	School     School    `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	TeacherID  uuid.UUID `gorm:"type:uuid" json:"teacher_id"`
	Teacher    User      `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	GradeLevel string    `json:"grade_level"`
	Subject    string    `json:"subject"`
	Capacity   int       `json:"capacity"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Classroom) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

// Enrollment represents student-classroom relationships
type Enrollment struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ClassroomID uuid.UUID `gorm:"type:uuid" json:"classroom_id"`
	Classroom   Classroom `gorm:"foreignKey:ClassroomID" json:"classroom,omitempty"`
	UserID      uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EnrolledAt  time.Time `json:"enrolled_at"`
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, withdrawn
}

func (e *Enrollment) BeforeCreate(tx *gorm.DB) error {
	e.ID = uuid.New()
	return nil
}

// Session represents app usage sessions
type Session struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Application     string    `gorm:"type:varchar(20);not null" json:"application"` // whiteboard, notebook
	ClassroomID     *uuid.UUID `gorm:"type:uuid" json:"classroom_id"`
	Classroom       *Classroom `gorm:"foreignKey:ClassroomID" json:"classroom,omitempty"`
	StartTime       time.Time `gorm:"not null" json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	DurationSeconds *int      `json:"duration_seconds"`
	DeviceType      string    `json:"device_type"`
	AppVersion      string    `json:"app_version"`
	CreatedAt       time.Time `json:"created_at"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}

// Event represents user interactions and system events
type Event struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	EventType   string    `gorm:"not null" json:"event_type"`
	UserID      uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SessionID   uuid.UUID `gorm:"type:uuid" json:"session_id"`
	Session     Session   `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	ClassroomID *uuid.UUID `gorm:"type:uuid" json:"classroom_id"`
	Classroom   *Classroom `gorm:"foreignKey:ClassroomID" json:"classroom,omitempty"`
	Timestamp   time.Time `gorm:"not null" json:"timestamp"`
	Application string    `gorm:"type:varchar(20)" json:"application"`
	Payload     JSONB     `gorm:"type:jsonb" json:"payload"`
	Metadata    JSONB     `gorm:"type:jsonb" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	e.ID = uuid.New()
	return nil
}

// Quiz represents quiz content and metadata
type Quiz struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Title            string    `gorm:"not null" json:"title"`
	ClassroomID      uuid.UUID `gorm:"type:uuid" json:"classroom_id"`
	Classroom        Classroom `gorm:"foreignKey:ClassroomID" json:"classroom,omitempty"`
	TeacherID        uuid.UUID `gorm:"type:uuid" json:"teacher_id"`
	Teacher          User      `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	QuestionCount    int       `json:"question_count"`
	TotalPoints      float64   `gorm:"type:decimal(10,2)" json:"total_points"`
	TimeLimitMinutes *int      `json:"time_limit_minutes"`
	CreatedAt        time.Time `json:"created_at"`
	PublishedAt      *time.Time `json:"published_at"`
	Status           string    `gorm:"type:varchar(20);default:'draft'" json:"status"` // draft, published, completed, archived
}

func (q *Quiz) BeforeCreate(tx *gorm.DB) error {
	q.ID = uuid.New()
	return nil
}

// QuizQuestion represents individual quiz questions
type QuizQuestion struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	QuizID         uuid.UUID `gorm:"type:uuid" json:"quiz_id"`
	Quiz           Quiz      `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	QuestionText   string    `gorm:"type:text;not null" json:"question_text"`
	QuestionType   string    `gorm:"type:varchar(50)" json:"question_type"` // multiple_choice, true_false, short_answer, essay
	Options        JSONB     `gorm:"type:jsonb" json:"options"`
	CorrectAnswer  string    `json:"correct_answer"`
	Points         float64   `gorm:"type:decimal(5,2)" json:"points"`
	OrderIndex     int       `json:"order_index"`
	CreatedAt      time.Time `json:"created_at"`
}

func (qq *QuizQuestion) BeforeCreate(tx *gorm.DB) error {
	qq.ID = uuid.New()
	return nil
}

// QuizResponse represents student answers to quiz questions
type QuizResponse struct {
	ID               uuid.UUID     `gorm:"type:uuid;primary_key" json:"id"`
	QuizID           uuid.UUID     `gorm:"type:uuid" json:"quiz_id"`
	Quiz             Quiz          `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	QuestionID       uuid.UUID     `gorm:"type:uuid" json:"question_id"`
	Question         QuizQuestion  `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	StudentID        uuid.UUID     `gorm:"type:uuid" json:"student_id"`
	Student          User          `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Answer           string        `json:"answer"`
	IsCorrect        *bool         `json:"is_correct"`
	PointsEarned     *float64      `gorm:"type:decimal(5,2)" json:"points_earned"`
	TimeTakenSeconds *int          `json:"time_taken_seconds"`
	SubmittedAt      *time.Time    `json:"submitted_at"`
	CreatedAt        time.Time     `json:"created_at"`
}

func (qr *QuizResponse) BeforeCreate(tx *gorm.DB) error {
	qr.ID = uuid.New()
	return nil
}

// DailyUserStats represents aggregated daily user statistics
type DailyUserStats struct {
	ID                      uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID                  uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User                    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Date                    time.Time `gorm:"type:date;not null" json:"date"`
	Application             string    `gorm:"type:varchar(20)" json:"application"`
	TotalSessionTime        int       `json:"total_session_time"`
	SessionCount            int       `json:"session_count"`
	EventCount              int       `json:"event_count"`
	QuizParticipationCount  int       `json:"quiz_participation_count"`
	AverageScore            *float64  `gorm:"type:decimal(5,2)" json:"average_score"`
	CreatedAt               time.Time `json:"created_at"`
}

func (dus *DailyUserStats) BeforeCreate(tx *gorm.DB) error {
	dus.ID = uuid.New()
	return nil
}

// ClassroomAnalytics represents aggregated classroom metrics
type ClassroomAnalytics struct {
	ID                     uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ClassroomID            uuid.UUID `gorm:"type:uuid" json:"classroom_id"`
	Classroom              Classroom `gorm:"foreignKey:ClassroomID" json:"classroom,omitempty"`
	Date                   time.Time `gorm:"type:date;not null" json:"date"`
	ActiveStudents         int       `json:"active_students"`
	QuizCount              int       `json:"quiz_count"`
	AverageEngagementScore *float64  `gorm:"type:decimal(5,2)" json:"average_engagement_score"`
	TotalContentCreated    int       `json:"total_content_created"`
	CreatedAt              time.Time `json:"created_at"`
}

func (ca *ClassroomAnalytics) BeforeCreate(tx *gorm.DB) error {
	ca.ID = uuid.New()
	return nil
}