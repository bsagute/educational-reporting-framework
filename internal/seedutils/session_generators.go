package seedutils

import (
	"fmt"
	"math/rand"
	"time"

	"reporting-framework/internal/models"

	"gorm.io/gorm"
)

// SessionGenerator handles creation of realistic user session data
type SessionGenerator struct {
	db     *gorm.DB
	logger Logger
	rand   *rand.Rand
}

// NewSessionGenerator creates a new session generator
func NewSessionGenerator(db *gorm.DB, logger Logger) *SessionGenerator {
	source := rand.NewSource(time.Now().UnixNano() + 4000)
	return &SessionGenerator{
		db:     db,
		logger: logger,
		rand:   rand.New(source),
	}
}

// SessionPattern defines realistic usage patterns for educational apps
type SessionPattern struct {
	MinDuration     int      // Minimum session duration in seconds
	MaxDuration     int      // Maximum session duration in seconds
	Applications    []string // Available applications
	DeviceTypes     []string // Common device types in education
	AppVersions     []string // Different app versions
	SessionsPerUser int      // Average sessions per user
	DaysBack        int      // How many days back to generate sessions
}

// GetEducationalSessionPattern returns realistic patterns for educational app usage
func GetEducationalSessionPattern() SessionPattern {
	return SessionPattern{
		MinDuration:     300,  // 5 minutes minimum
		MaxDuration:     3600, // 1 hour maximum
		Applications:    []string{"whiteboard", "notebook"},
		DeviceTypes:     []string{"tablet", "laptop", "chromebook", "desktop"},
		AppVersions:     []string{"2.0.1", "2.1.0", "2.1.1", "2.2.0"},
		SessionsPerUser: 15, // 15 sessions per user over the time period
		DaysBack:        30, // Generate sessions over last 30 days
	}
}

// GenerateSessionsForUsers creates realistic session data for a subset of users
func (sg *SessionGenerator) GenerateSessionsForUsers(users []models.User, pattern SessionPattern) ([]models.Session, error) {
	sg.logger.Info("Starting session generation", "userCount", len(users), "sessionsPerUser", pattern.SessionsPerUser)

	if len(users) == 0 {
		return nil, fmt.Errorf("cannot generate sessions: no users provided")
	}

	var allSessions []models.Session
	totalSessions := 0

	// Limit to first 200 users to avoid excessive data during development
	maxUsers := 200
	if len(users) > maxUsers {
		users = users[:maxUsers]
		sg.logger.Info("Limiting session generation to avoid excessive data", "limitedTo", maxUsers)
	}

	for i, user := range users {
		userSessions, err := sg.generateSessionsForUser(user, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to generate sessions for user %s: %w", user.Username, err)
		}

		allSessions = append(allSessions, userSessions...)
		totalSessions += len(userSessions)

		if (i+1)%50 == 0 {
			sg.logger.Debug("Session generation progress", "users_processed", i+1, "total_sessions", totalSessions)
		}
	}

	sg.logger.Info("Successfully generated all sessions", "totalSessions", totalSessions)
	return allSessions, nil
}

// generateSessionsForUser creates sessions for a single user
func (sg *SessionGenerator) generateSessionsForUser(user models.User, pattern SessionPattern) ([]models.Session, error) {
	var sessions []models.Session

	for i := 0; i < pattern.SessionsPerUser; i++ {
		session, err := sg.createRealisticSession(user, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to create session %d for user %s: %w", i+1, user.Username, err)
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// createRealisticSession generates a single realistic session
func (sg *SessionGenerator) createRealisticSession(user models.User, pattern SessionPattern) (models.Session, error) {
	// Generate random start time within the specified range
	daysAgo := sg.rand.Intn(pattern.DaysBack)
	hour := 8 + sg.rand.Intn(10) // Sessions mostly during school hours (8 AM - 6 PM)
	minute := sg.rand.Intn(60)

	startTime := time.Now().
		AddDate(0, 0, -daysAgo).
		Truncate(24 * time.Hour). // Start of day
		Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

	// Generate session duration
	duration := pattern.MinDuration + sg.rand.Intn(pattern.MaxDuration-pattern.MinDuration)
	endTime := startTime.Add(time.Duration(duration) * time.Second)

	// Select random application and device
	application := pattern.Applications[sg.rand.Intn(len(pattern.Applications))]
	deviceType := pattern.DeviceTypes[sg.rand.Intn(len(pattern.DeviceTypes))]
	appVersion := pattern.AppVersions[sg.rand.Intn(len(pattern.AppVersions))]

	session := models.Session{
		UserID:          user.ID,
		Application:     application,
		StartTime:       startTime,
		EndTime:         &endTime,
		DurationSeconds: &duration,
		DeviceType:      deviceType,
		AppVersion:      appVersion,
	}

	if err := sg.db.Create(&session).Error; err != nil {
		return models.Session{}, fmt.Errorf("failed to save session to database: %w", err)
	}

	return session, nil
}

// EventGenerator creates realistic user interaction events
type EventGenerator struct {
	db     *gorm.DB
	logger Logger
	rand   *rand.Rand
}

// NewEventGenerator creates a new event generator
func NewEventGenerator(db *gorm.DB, logger Logger) *EventGenerator {
	source := rand.NewSource(time.Now().UnixNano() + 5000)
	return &EventGenerator{
		db:     db,
		logger: logger,
		rand:   rand.New(source),
	}
}

// EventPattern defines realistic event generation patterns
type EventPattern struct {
	EventTypes       []string
	MinEventsPerSession int
	MaxEventsPerSession int
}

// GetEducationalEventPattern returns realistic event patterns for educational apps
func GetEducationalEventPattern() EventPattern {
	return EventPattern{
		EventTypes: []string{
			"page_view",
			"quiz_answer_submitted",
			"content_created",
			"note_taken",
			"drawing_created",
			"file_uploaded",
			"collaborative_edit",
			"video_watched",
			"assignment_submitted",
		},
		MinEventsPerSession: 5,
		MaxEventsPerSession: 25,
	}
}

// GenerateEventsForSessions creates realistic events for existing sessions
func (eg *EventGenerator) GenerateEventsForSessions(sessions []models.Session, pattern EventPattern) error {
	eg.logger.Info("Starting event generation for sessions", "sessionCount", len(sessions))

	totalEvents := 0

	for i, session := range sessions {
		events, err := eg.generateEventsForSession(session, pattern)
		if err != nil {
			return fmt.Errorf("failed to generate events for session %s: %w", session.ID, err)
		}

		totalEvents += events

		if (i+1)%100 == 0 {
			eg.logger.Debug("Event generation progress", "sessions_processed", i+1, "total_events", totalEvents)
		}
	}

	eg.logger.Info("Successfully generated events for all sessions", "totalEvents", totalEvents)
	return nil
}

// generateEventsForSession creates events for a single session
func (eg *EventGenerator) generateEventsForSession(session models.Session, pattern EventPattern) (int, error) {
	if session.EndTime == nil || session.DurationSeconds == nil {
		// Skip incomplete sessions
		return 0, nil
	}

	eventCount := pattern.MinEventsPerSession + eg.rand.Intn(pattern.MaxEventsPerSession-pattern.MinEventsPerSession+1)
	sessionDuration := time.Duration(*session.DurationSeconds) * time.Second

	for i := 0; i < eventCount; i++ {
		// Generate event time within session duration
		eventOffset := time.Duration(eg.rand.Int63n(int64(sessionDuration)))
		eventTime := session.StartTime.Add(eventOffset)

		eventType := pattern.EventTypes[eg.rand.Intn(len(pattern.EventTypes))]

		event := models.Event{
			EventType:   eventType,
			UserID:      session.UserID,
			SessionID:   session.ID,
			Timestamp:   eventTime,
			Application: session.Application,
			Payload:     eg.generateEventPayload(eventType, session.Application),
			Metadata:    eg.generateEventMetadata(session),
		}

		if err := eg.db.Create(&event).Error; err != nil {
			return i, fmt.Errorf("failed to create event %d: %w", i+1, err)
		}
	}

	return eventCount, nil
}

// generateEventPayload creates realistic payload data based on event type
func (eg *EventGenerator) generateEventPayload(eventType, application string) models.JSONB {
	basePayload := models.JSONB{
		"timestamp": time.Now().Unix(),
		"application": application,
	}

	switch eventType {
	case "page_view":
		basePayload["page_id"] = fmt.Sprintf("page_%d", eg.rand.Intn(100))
		basePayload["view_duration"] = 30 + eg.rand.Intn(300) // 30 seconds to 5 minutes

	case "quiz_answer_submitted":
		basePayload["quiz_id"] = fmt.Sprintf("quiz_%d", eg.rand.Intn(50))
		basePayload["question_id"] = fmt.Sprintf("q_%d", eg.rand.Intn(10))
		basePayload["answer"] = []string{"A", "B", "C", "D"}[eg.rand.Intn(4)]
		basePayload["time_taken_seconds"] = 15 + eg.rand.Intn(120)

	case "content_created":
		contentTypes := []string{"note", "drawing", "document", "presentation"}
		basePayload["content_type"] = contentTypes[eg.rand.Intn(len(contentTypes))]
		basePayload["content_size"] = 100 + eg.rand.Intn(10000) // Size in characters/bytes

	case "note_taken":
		basePayload["note_length"] = 50 + eg.rand.Intn(500)
		basePayload["note_category"] = []string{"lesson", "homework", "reminder", "idea"}[eg.rand.Intn(4)]

	case "drawing_created":
		basePayload["drawing_tools_used"] = eg.rand.Intn(5) + 1
		basePayload["drawing_time"] = 60 + eg.rand.Intn(600) // 1-10 minutes

	case "video_watched":
		basePayload["video_id"] = fmt.Sprintf("video_%d", eg.rand.Intn(200))
		basePayload["watch_duration"] = 30 + eg.rand.Intn(1800) // 30 seconds to 30 minutes
		basePayload["completion_percentage"] = eg.rand.Float64() * 100

	case "collaborative_edit":
		basePayload["document_id"] = fmt.Sprintf("doc_%d", eg.rand.Intn(100))
		basePayload["edit_type"] = []string{"text_added", "text_deleted", "formatting", "comment"}[eg.rand.Intn(4)]
		basePayload["collaborators"] = eg.rand.Intn(5) + 1

	default:
		basePayload["action"] = "generic_interaction"
		basePayload["duration"] = eg.rand.Intn(300)
	}

	return basePayload
}

// generateEventMetadata creates realistic metadata for events
func (eg *EventGenerator) generateEventMetadata(session models.Session) models.JSONB {
	userAgents := []string{
		"Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
		"Mozilla/5.0 (X11; CrOS armv7l 13904.77.0) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	ipRanges := []string{
		"192.168.1.", "10.0.0.", "172.16.1.", "192.168.100.",
	}

	selectedIP := ipRanges[eg.rand.Intn(len(ipRanges))] + fmt.Sprintf("%d", 1+eg.rand.Intn(254))

	return models.JSONB{
		"user_agent":      userAgents[eg.rand.Intn(len(userAgents))],
		"ip_address":      selectedIP,
		"device_type":     session.DeviceType,
		"app_version":     session.AppVersion,
		"screen_size":     fmt.Sprintf("%dx%d", 1024+eg.rand.Intn(1920), 768+eg.rand.Intn(1080)),
		"connection_type": []string{"wifi", "ethernet", "cellular"}[eg.rand.Intn(3)],
		"location":        "school_network",
	}
}