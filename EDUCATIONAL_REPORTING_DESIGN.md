# Educational Applications Reporting Framework - Technical Design

## Executive Summary

This document outlines the design and implementation of a comprehensive reporting framework for educational applications ecosystem consisting of:
- **Whiteboard App**: Digital classroom whiteboard for teachers
- **Notebook App**: Personal note-taking and content consumption for students/teachers
- **Quiz Feature**: Synchronized quiz functionality between both apps

The framework handles data from approximately:
- 1,000 schools
- 30 classrooms per school (30,000 total)
- 30 students per classroom (900,000 total users)

## 1. Data Collection Strategy

### 1.1 Event-Based Tracking Methodology

The reporting system uses event-driven architecture to capture user interactions and system events in real-time.

#### Core Event Categories:

**Session Events:**
- `session_start`: User opens application
- `session_end`: User closes application
- `session_heartbeat`: Periodic session activity check

**Content Events:**
- `content_created`: New content (notes, drawings, documents)
- `content_viewed`: Content accessed/opened
- `content_modified`: Content edited
- `content_shared`: Content shared between users

**Quiz Events:**
- `quiz_created`: Teacher creates new quiz
- `quiz_started`: Quiz session begins
- `quiz_question_displayed`: Question shown to students
- `quiz_answer_submitted`: Student submits answer
- `quiz_completed`: Quiz session ends
- `quiz_results_viewed`: Results accessed

**Interaction Events:**
- `whiteboard_draw`: Drawing/writing on whiteboard
- `whiteboard_erase`: Content erased
- `notebook_write`: Notes written
- `file_upload`: Files uploaded to system
- `sync_initiated`: Cross-app synchronization

#### Event Data Structure:
```json
{
  "event_id": "uuid",
  "event_type": "quiz_answer_submitted",
  "timestamp": "2024-01-15T10:30:45Z",
  "user_id": "user_123",
  "session_id": "session_456",
  "school_id": "school_789",
  "classroom_id": "class_101",
  "application": "notebook", // or "whiteboard"
  "metadata": {
    "quiz_id": "quiz_789",
    "question_id": "q_123",
    "answer": "option_b",
    "time_spent": 45,
    "attempt_number": 1
  },
  "device_info": {
    "platform": "android",
    "version": "10.0",
    "device_model": "Samsung Galaxy Tab"
  }
}
```

### 1.2 Metrics to Track

**Student Performance Metrics:**
- Quiz scores and accuracy rates
- Time spent on questions
- Answer patterns and learning progression
- Content engagement duration
- Collaboration frequency

**Classroom Engagement Metrics:**
- Active participation rates
- Real-time sync usage
- Content sharing frequency
- Session duration and frequency
- Device usage patterns

**Content Effectiveness Metrics:**
- Content creation vs consumption ratios
- Popular content types and topics
- Content revision patterns
- Cross-app content usage

## 2. Database Schema Design

### 2.1 Core Entities

**Schools & Organization:**
```sql
-- Schools table
CREATE TABLE schools (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    district VARCHAR(255),
    region VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Classrooms table
CREATE TABLE classrooms (
    id UUID PRIMARY KEY,
    school_id UUID REFERENCES schools(id),
    name VARCHAR(255) NOT NULL,
    grade_level INTEGER,
    subject VARCHAR(100),
    teacher_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Users & Sessions:**
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    school_id UUID REFERENCES schools(id),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) CHECK (role IN ('teacher', 'student', 'admin')),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP
);

-- User classroom associations
CREATE TABLE user_classrooms (
    user_id UUID REFERENCES users(id),
    classroom_id UUID REFERENCES classrooms(id),
    role VARCHAR(50) CHECK (role IN ('teacher', 'student')),
    enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, classroom_id)
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    classroom_id UUID REFERENCES classrooms(id),
    application VARCHAR(50) CHECK (application IN ('whiteboard', 'notebook')),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    device_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Content & Quizzes:**
```sql
-- Content table
CREATE TABLE content (
    id UUID PRIMARY KEY,
    creator_id UUID REFERENCES users(id),
    classroom_id UUID REFERENCES classrooms(id),
    title VARCHAR(255),
    content_type VARCHAR(50) CHECK (content_type IN ('note', 'drawing', 'document', 'quiz')),
    content_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_shared BOOLEAN DEFAULT FALSE
);

-- Quizzes table
CREATE TABLE quizzes (
    id UUID PRIMARY KEY,
    creator_id UUID REFERENCES users(id),
    classroom_id UUID REFERENCES classrooms(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    total_questions INTEGER,
    time_limit_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT FALSE
);

-- Quiz questions table
CREATE TABLE quiz_questions (
    id UUID PRIMARY KEY,
    quiz_id UUID REFERENCES quizzes(id),
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) CHECK (question_type IN ('multiple_choice', 'true_false', 'short_answer')),
    options JSONB, -- For multiple choice questions
    correct_answer TEXT,
    points INTEGER DEFAULT 1,
    order_index INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Events & Analytics:**
```sql
-- Events table (main event store)
CREATE TABLE events (
    id UUID PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES sessions(id),
    classroom_id UUID REFERENCES classrooms(id),
    school_id UUID REFERENCES schools(id),
    application VARCHAR(50),
    timestamp TIMESTAMP NOT NULL,
    metadata JSONB,
    device_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quiz submissions table
CREATE TABLE quiz_submissions (
    id UUID PRIMARY KEY,
    quiz_id UUID REFERENCES quizzes(id),
    student_id UUID REFERENCES users(id),
    question_id UUID REFERENCES quiz_questions(id),
    submitted_answer TEXT,
    is_correct BOOLEAN,
    time_spent_seconds INTEGER,
    attempt_number INTEGER DEFAULT 1,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Aggregated metrics tables for performance
CREATE TABLE daily_user_metrics (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    date DATE NOT NULL,
    session_count INTEGER DEFAULT 0,
    total_session_duration_seconds INTEGER DEFAULT 0,
    events_count INTEGER DEFAULT 0,
    quiz_attempts INTEGER DEFAULT 0,
    avg_quiz_score DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE TABLE daily_classroom_metrics (
    id UUID PRIMARY KEY,
    classroom_id UUID REFERENCES classrooms(id),
    date DATE NOT NULL,
    active_students_count INTEGER DEFAULT 0,
    total_quiz_sessions INTEGER DEFAULT 0,
    avg_engagement_score DECIMAL(5,2),
    content_created_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(classroom_id, date)
);
```

### 2.2 Indexes for Performance
```sql
-- Performance indexes
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_user_id ON events(user_id);
CREATE INDEX idx_events_classroom_id ON events(classroom_id);
CREATE INDEX idx_events_type_timestamp ON events(event_type, timestamp);
CREATE INDEX idx_sessions_user_time ON sessions(user_id, start_time);
CREATE INDEX idx_quiz_submissions_quiz_student ON quiz_submissions(quiz_id, student_id);
CREATE INDEX idx_daily_metrics_date ON daily_user_metrics(date);
CREATE INDEX idx_classroom_metrics_date ON daily_classroom_metrics(date);
```

## 3. API Design

### 3.1 Data Ingestion Endpoints

**Event Ingestion API:**
```
POST /api/v1/events
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "events": [
    {
      "event_type": "quiz_answer_submitted",
      "timestamp": "2024-01-15T10:30:45Z",
      "metadata": {
        "quiz_id": "quiz_789",
        "question_id": "q_123",
        "answer": "option_b",
        "time_spent": 45
      }
    }
  ]
}
```

**Batch Session Data:**
```
POST /api/v1/sessions/batch
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "sessions": [
    {
      "application": "whiteboard",
      "start_time": "2024-01-15T09:00:00Z",
      "end_time": "2024-01-15T10:30:00Z",
      "classroom_id": "class_101",
      "events": [...] // Associated events
    }
  ]
}
```

### 3.2 Report Generation Endpoints

**Student Performance Reports:**
```
GET /api/v1/reports/student-performance
Query Parameters:
- student_id: UUID
- classroom_id: UUID (optional)
- date_from: Date
- date_to: Date
- include_details: boolean

Response:
{
  "student_id": "user_123",
  "period": {"from": "2024-01-01", "to": "2024-01-31"},
  "overall_stats": {
    "avg_quiz_score": 85.5,
    "total_quizzes": 15,
    "total_session_time": 7200,
    "engagement_score": 92.3
  },
  "quiz_performance": [...],
  "learning_progression": [...]
}
```

**Classroom Engagement Reports:**
```
GET /api/v1/reports/classroom-engagement
Query Parameters:
- classroom_id: UUID
- date_from: Date
- date_to: Date
- metrics: string[] (optional)

Response:
{
  "classroom_id": "class_101",
  "engagement_metrics": {
    "active_participation_rate": 87.5,
    "avg_session_duration": 45.2,
    "collaboration_events": 234,
    "content_sharing_frequency": 12.3
  },
  "student_breakdown": [...],
  "timeline_data": [...]
}
```

**Content Effectiveness Reports:**
```
GET /api/v1/reports/content-effectiveness
Query Parameters:
- school_id: UUID (optional)
- classroom_id: UUID (optional)
- content_type: string (optional)
- date_from: Date
- date_to: Date

Response:
{
  "content_analytics": {
    "most_engaging_content": [...],
    "content_usage_patterns": {...},
    "effectiveness_scores": [...]
  },
  "recommendations": [...]
}
```

### 3.3 Generic Query API (Cube.dev style)

**Cube-style Query Endpoint:**
```
POST /api/v1/query
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "measures": ["events.count", "users.avg_session_duration"],
  "dimensions": ["users.role", "events.event_type", "time.date"],
  "timeDimensions": [
    {
      "dimension": "events.timestamp",
      "granularity": "day",
      "dateRange": ["2024-01-01", "2024-01-31"]
    }
  ],
  "filters": [
    {
      "member": "classrooms.school_id",
      "operator": "equals",
      "values": ["school_123"]
    }
  ],
  "order": [["events.timestamp", "desc"]]
}
```

### 3.4 Authentication & Authorization

**JWT Token Structure:**
```json
{
  "sub": "user_123",
  "school_id": "school_789",
  "role": "teacher",
  "permissions": ["read:reports", "write:events"],
  "iat": 1642248000,
  "exp": 1642251600
}
```

**Role-based Access Control:**
- **Students**: Can only submit events for themselves, view their own performance
- **Teachers**: Can view reports for their classrooms, submit events for their sessions
- **School Admins**: Can view all reports for their school
- **System Admins**: Full access across all schools

## 4. Implementation Architecture

The reporting framework will be implemented using the existing gRPC microservice template with the following services:

### 4.1 Service Architecture
- **Event Ingestion Service**: Handles real-time event collection
- **Analytics Service**: Processes and aggregates data
- **Report Generation Service**: Creates reports and dashboards
- **Query Service**: Handles generic cube-style queries

### 4.2 Technology Stack
- **Database**: PostgreSQL with time-series optimizations
- **Message Queue**: Redis for event buffering
- **Caching**: Redis for report caching
- **Monitoring**: OpenTelemetry for observability

This framework provides scalable, real-time analytics capabilities for the educational ecosystem with comprehensive reporting and insights generation.