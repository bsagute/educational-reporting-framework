# Educational Applications Reporting Framework

## 🎯 Project Overview

A comprehensive reporting framework designed for educational applications ecosystem consisting of **Whiteboard App** (digital classroom whiteboard for teachers) and **Notebook App** (personal note-taking for students/teachers). The framework captures, stores, and analyzes user interactions with a particular focus on the synchronized **Quiz feature**.

### 📊 Scale Requirements
- **1,000 schools**
- **30 classrooms per school** (30,000 total)
- **30 students per classroom** (900,000 total users)

---

## 🏗️ System Architecture

### High-Level Architecture Diagram

```plantuml
@startuml Educational_Reporting_Architecture
!define RECTANGLE class

title Educational Applications Reporting Framework - High Level Architecture

cloud "Educational Apps Ecosystem" {
  [Whiteboard App\n(Teachers)] as WA
  [Notebook App\n(Students/Teachers)] as NA

  WA <--> NA : Quiz Sync\nReal-time
}

package "Reporting Framework" {
  [API Gateway] as API
  [Event Ingestion Service] as EIS
  [Analytics Service] as AS
  [Report Generation Service] as RGS
  [Query Service] as QS
}

database "PostgreSQL\nTime-Series Optimized" as DB {
  [Core Tables] as CT
  [Aggregated Metrics] as AM
  [Materialized Views] as MV
}

cloud "External Services" {
  [Redis Cache] as Redis
  [Message Queue] as MQ
  [Monitoring] as MON
}

WA --> API : Event Data\n(REST/gRPC)
NA --> API : Event Data\n(REST/gRPC)

API --> EIS : Route Events
EIS --> DB : Store Events
EIS --> Redis : Cache State
EIS --> MQ : Queue Jobs

MQ --> AS : Process Aggregations
AS --> DB : Update Metrics
AS --> MV : Refresh Views

API --> RGS : Generate Reports
RGS --> DB : Query Data
RGS --> Redis : Cache Reports

API --> QS : Cube.dev Queries
QS --> DB : Execute SQL

DB --> Redis : Cache Results
Redis --> API : Fast Responses

MON --> EIS : Monitor Events
MON --> AS : Monitor Jobs
MON --> RGS : Monitor Reports

@enduml
```

### Component Architecture Diagram

```plantuml
@startuml Component_Architecture
!define RECTANGLE class

title Educational Reporting Framework - Component Architecture

package "Client Applications" {
  component [Whiteboard App] as WB
  component [Notebook App] as NB
}

package "API Layer" {
  component [Event Ingestion API] as EIA
  component [Reports API] as RA
  component [Analytics API] as AA
  component [Generic Query API] as GQA
  component [Admin API] as ADA
}

package "Service Layer" {
  component [Event Processing Service] as EPS
  component [Aggregation Service] as AGS
  component [Report Generation Service] as RGS
  component [Query Builder Service] as QBS
}

package "Data Layer" {
  database [PostgreSQL] as PG {
    [Events Table]
    [Users Table]
    [Schools Table]
    [Classrooms Table]
    [Quiz Data Tables]
    [Aggregated Metrics]
    [Materialized Views]
  }

  database [Redis Cache] as RC {
    [Session Cache]
    [Report Cache]
    [Query Cache]
  }
}

package "Background Jobs" {
  component [Metrics Aggregator] as MA
  component [Report Scheduler] as RS
  component [Data Cleanup] as DC
}

WB --> EIA : POST /events
NB --> EIA : POST /events
WB --> RA : GET /reports
NB --> RA : GET /reports

EIA --> EPS : Process Events
RA --> RGS : Generate Reports
AA --> AGS : Real-time Analytics
GQA --> QBS : Build Queries
ADA --> EPS : Admin Operations

EPS --> PG : Store Events
EPS --> RC : Cache State
RGS --> PG : Query Data
RGS --> RC : Cache Reports
QBS --> PG : Execute Queries
AGS --> PG : Read Metrics

MA --> PG : Update Aggregations
RS --> RGS : Schedule Reports
DC --> PG : Cleanup Old Data

@enduml
```

---

## 🗄️ Database Schema Design

### Entity Relationship Diagram

```plantuml
@startuml Educational_Reporting_ERD
!define TABLE entity

title Educational Reporting Framework - Entity Relationship Diagram

entity "schools" as schools {
  * id : UUID <<PK>>
  --
  * name : VARCHAR(255)
  district : VARCHAR(255)
  region : VARCHAR(100)
  contact_email : VARCHAR(255)
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "classrooms" as classrooms {
  * id : UUID <<PK>>
  --
  * school_id : UUID <<FK>>
  * name : VARCHAR(255)
  grade_level : INTEGER
  subject : VARCHAR(100)
  teacher_id : UUID <<FK>>
  max_students : INTEGER
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "users" as users {
  * id : UUID <<PK>>
  --
  * school_id : UUID <<FK>>
  * username : VARCHAR(100) <<UK>>
  * role : VARCHAR(50)
  email : VARCHAR(255)
  first_name : VARCHAR(100)
  last_name : VARCHAR(100)
  last_active : TIMESTAMP
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "user_classrooms" as user_classrooms {
  * user_id : UUID <<PK,FK>>
  * classroom_id : UUID <<PK,FK>>
  --
  * role : VARCHAR(50)
  enrolled_at : TIMESTAMP
  is_active : BOOLEAN
}

entity "sessions" as sessions {
  * id : UUID <<PK>>
  --
  * user_id : UUID <<FK>>
  classroom_id : UUID <<FK>>
  * application : VARCHAR(50)
  * start_time : TIMESTAMP
  end_time : TIMESTAMP
  duration_seconds : INTEGER
  device_info : JSONB
  ip_address : INET
  created_at : TIMESTAMP
}

entity "events" as events {
  * id : UUID <<PK>>
  --
  * event_type : VARCHAR(100)
  user_id : UUID <<FK>>
  session_id : UUID <<FK>>
  classroom_id : UUID <<FK>>
  school_id : UUID <<FK>>
  application : VARCHAR(50)
  * timestamp : TIMESTAMP
  metadata : JSONB
  device_info : JSONB
  created_at : TIMESTAMP
}

entity "quizzes" as quizzes {
  * id : UUID <<PK>>
  --
  * creator_id : UUID <<FK>>
  * classroom_id : UUID <<FK>>
  * title : VARCHAR(255)
  description : TEXT
  total_questions : INTEGER
  time_limit_minutes : INTEGER
  max_attempts : INTEGER
  is_active : BOOLEAN
  start_time : TIMESTAMP
  end_time : TIMESTAMP
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "quiz_questions" as quiz_questions {
  * id : UUID <<PK>>
  --
  * quiz_id : UUID <<FK>>
  * question_text : TEXT
  * question_type : VARCHAR(50)
  * order_index : INTEGER
  options : JSONB
  correct_answer : TEXT
  points : INTEGER
  explanation : TEXT
  created_at : TIMESTAMP
}

entity "quiz_sessions" as quiz_sessions {
  * id : UUID <<PK>>
  --
  * quiz_id : UUID <<FK>>
  * student_id : UUID <<FK>>
  started_at : TIMESTAMP
  completed_at : TIMESTAMP
  total_score : INTEGER
  max_possible_score : INTEGER
  percentage_score : DECIMAL(5,2)
  time_spent_seconds : INTEGER
  attempt_number : INTEGER
  is_completed : BOOLEAN
}

entity "quiz_submissions" as quiz_submissions {
  * id : UUID <<PK>>
  --
  * quiz_id : UUID <<FK>>
  * student_id : UUID <<FK>>
  * question_id : UUID <<FK>>
  submitted_answer : TEXT
  is_correct : BOOLEAN
  points_earned : INTEGER
  time_spent_seconds : INTEGER
  attempt_number : INTEGER
  submitted_at : TIMESTAMP
}

entity "content" as content {
  * id : UUID <<PK>>
  --
  * creator_id : UUID <<FK>>
  classroom_id : UUID <<FK>>
  title : VARCHAR(255)
  * content_type : VARCHAR(50)
  content_data : JSONB
  file_size_bytes : BIGINT
  is_shared : BOOLEAN
  share_permissions : JSONB
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

' Relationships
schools ||--o{ classrooms : "has many"
schools ||--o{ users : "has many"
classrooms ||--o{ user_classrooms : "has many"
users ||--o{ user_classrooms : "has many"
users ||--o{ sessions : "creates"
users ||--o{ events : "generates"
users ||--o{ quizzes : "creates"
users ||--o{ quiz_sessions : "participates"
users ||--o{ content : "creates"
classrooms ||--o{ sessions : "hosts"
classrooms ||--o{ events : "contains"
classrooms ||--o{ quizzes : "contains"
sessions ||--o{ events : "contains"
quizzes ||--o{ quiz_questions : "contains"
quizzes ||--o{ quiz_sessions : "has attempts"
quiz_questions ||--o{ quiz_submissions : "receives answers"
quiz_sessions ||--o{ quiz_submissions : "contains"
users ||--o{ classrooms : "teaches"

@enduml
```

### Aggregated Metrics Schema

```plantuml
@startuml Aggregated_Metrics_Schema
!define TABLE entity

title Aggregated Metrics and Performance Tables

entity "daily_user_metrics" as dum {
  * id : UUID <<PK>>
  --
  * user_id : UUID <<FK>>
  * school_id : UUID <<FK>>
  * date : DATE <<UK>>
  session_count : INTEGER
  total_session_duration_seconds : INTEGER
  avg_session_duration_seconds : DECIMAL(10,2)
  events_count : INTEGER
  quiz_attempts : INTEGER
  quiz_completions : INTEGER
  avg_quiz_score : DECIMAL(5,2)
  content_created_count : INTEGER
  content_viewed_count : INTEGER
  whiteboard_events : INTEGER
  notebook_events : INTEGER
  collaboration_events : INTEGER
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "daily_classroom_metrics" as dcm {
  * id : UUID <<PK>>
  --
  * classroom_id : UUID <<FK>>
  * school_id : UUID <<FK>>
  * date : DATE <<UK>>
  total_students : INTEGER
  active_students_count : INTEGER
  participation_rate : DECIMAL(5,2)
  total_sessions : INTEGER
  avg_session_duration_minutes : DECIMAL(10,2)
  total_quiz_sessions : INTEGER
  avg_quiz_completion_rate : DECIMAL(5,2)
  avg_class_quiz_score : DECIMAL(5,2)
  content_created_count : INTEGER
  content_shared_count : INTEGER
  whiteboard_usage_minutes : INTEGER
  notebook_usage_minutes : INTEGER
  sync_events_count : INTEGER
  engagement_score : DECIMAL(5,2)
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "weekly_school_metrics" as wsm {
  * id : UUID <<PK>>
  --
  * school_id : UUID <<FK>>
  * week_start_date : DATE <<UK>>
  total_classrooms : INTEGER
  active_classrooms : INTEGER
  total_users : INTEGER
  active_users : INTEGER
  total_students : INTEGER
  active_students : INTEGER
  total_teachers : INTEGER
  active_teachers : INTEGER
  total_sessions : INTEGER
  avg_daily_sessions : DECIMAL(10,2)
  total_quiz_sessions : INTEGER
  avg_school_engagement : DECIMAL(5,2)
  total_content_created : INTEGER
  platform_adoption_rate : DECIMAL(5,2)
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "content_metrics" as cm {
  * id : UUID <<PK>>
  --
  * content_id : UUID <<FK>> <<UK>>
  classroom_id : UUID <<FK>>
  school_id : UUID <<FK>>
  * content_type : VARCHAR(50)
  view_count : INTEGER
  unique_viewers : INTEGER
  avg_view_duration_seconds : DECIMAL(10,2)
  interaction_count : INTEGER
  share_count : INTEGER
  effectiveness_score : DECIMAL(5,2)
  last_viewed_at : TIMESTAMP
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "quiz_analytics" as qa {
  * id : UUID <<PK>>
  --
  * quiz_id : UUID <<FK>> <<UK>>
  * classroom_id : UUID <<FK>>
  total_attempts : INTEGER
  unique_participants : INTEGER
  completion_rate : DECIMAL(5,2)
  avg_score : DECIMAL(5,2)
  avg_time_spent_minutes : DECIMAL(10,2)
  difficulty_score : DECIMAL(5,2)
  question_analytics : JSONB
  last_attempt_at : TIMESTAMP
  created_at : TIMESTAMP
  updated_at : TIMESTAMP
}

entity "active_sessions" as as {
  * id : UUID <<PK>>
  --
  * session_id : UUID <<FK>> <<UK>>
  * user_id : UUID <<FK>>
  classroom_id : UUID <<FK>>
  * application : VARCHAR(50)
  last_heartbeat : TIMESTAMP
  current_activity : VARCHAR(100)
  activity_metadata : JSONB
  created_at : TIMESTAMP
}

entity "hourly_event_aggregates" as hea {
  * id : UUID <<PK>>
  --
  * hour_timestamp : TIMESTAMP <<UK>>
  school_id : UUID <<FK>>
  classroom_id : UUID <<FK>>
  * event_type : VARCHAR(100) <<UK>>
  application : VARCHAR(50) <<UK>>
  event_count : INTEGER
  unique_users : INTEGER
  created_at : TIMESTAMP
}

' Relationships to main tables
dum }o--|| users : "metrics for"
dcm }o--|| classrooms : "metrics for"
wsm }o--|| schools : "metrics for"
cm }o--|| content : "metrics for"
qa }o--|| quizzes : "analytics for"
as }o--|| sessions : "tracks"
hea }o--|| schools : "aggregates for"
hea }o--|| classrooms : "aggregates for"

@enduml
```

---

## 🔄 Sequence Diagrams

### Real-time Event Ingestion Workflow

```plantuml
@startuml Event_Ingestion_Sequence
title Real-time Event Ingestion Workflow

participant "Whiteboard App" as WA
participant "Notebook App" as NA
participant "API Gateway" as API
participant "Event Service" as ES
participant "PostgreSQL" as DB
participant "Redis Cache" as Redis
participant "Message Queue" as MQ

note over WA, NA : User interactions generate events

WA -> API : POST /api/v1/events\n(batch events)
note right of WA : Events: whiteboard_draw,\ncontent_created, etc.

API -> ES : Validate & route events
ES -> ES : Authenticate user
ES -> ES : Enrich with context\n(school_id, session_id)

par Parallel Processing
    ES -> DB : Batch insert events
    ES -> Redis : Cache session state
    ES -> MQ : Queue aggregation jobs
end

ES -> API : 201 Created (event_ids)
API -> WA : Response

note over NA : Student submits quiz answer
NA -> API : POST /api/v1/events\n(quiz_answer_submitted)
API -> ES : Process quiz submission
ES -> ES : Validate quiz data
ES -> DB : Store quiz submission
ES -> ES : Calculate real-time scores
ES -> Redis : Update live analytics
ES -> API : 201 Created
API -> NA : Response

note over MQ : Background aggregation
MQ -> ES : Process aggregation job
ES -> DB : Update daily_user_metrics
ES -> DB : Update daily_classroom_metrics

@enduml
```

### Student Performance Report Generation

```plantuml
@startuml Student_Performance_Report
title Student Performance Report Generation

participant "Teacher Dashboard" as TD
participant "Reports API" as API
participant "Reports Service" as RS
participant "Redis Cache" as Cache
participant "PostgreSQL" as DB
participant "Analytics Engine" as AE

TD -> API : GET /api/v1/reports/student-performance\n?student_id=xyz&date_from=2024-01-01
API -> API : Authenticate teacher access

API -> Cache : Check cached report
alt Cache Hit
    Cache -> API : Return cached data
    API -> TD : 200 OK (cached report)
else Cache Miss
    API -> RS : Generate new report
    RS -> DB : Query daily_user_metrics
    RS -> DB : Query quiz_sessions
    RS -> DB : Query events for engagement

    RS -> AE : Calculate engagement score
    AE -> AE : Process learning progression
    AE -> RS : Calculated metrics

    RS -> RS : Format response
    RS -> Cache : Cache report (TTL: 1 hour)
    RS -> API : Performance report
    API -> TD : 200 OK (performance report)
end

note over TD : Teacher views detailed analytics
TD -> API : GET /api/v1/reports/student-performance\n?include_details=true
API -> RS : Get detailed report
RS -> DB : Query detailed quiz performance
RS -> DB : Query learning progression timeline
RS -> API : Detailed report
API -> TD : 200 OK (detailed report)

@enduml
```

### Quiz Session Workflow

```plantuml
@startuml Quiz_Session_Workflow
title Real-time Quiz Session Workflow

participant "Teacher\n(Whiteboard)" as T
participant "Students\n(Notebook)" as S
participant "API" as API
participant "PostgreSQL" as DB
participant "WebSocket" as WS
participant "Analytics" as A

note over T : Teacher creates quiz
T -> API : POST /api/v1/admin/quizzes
API -> DB : Store quiz and questions
API -> T : 201 Created (quiz_id)

T -> API : PUT /api/v1/admin/quizzes/{id}/activate
API -> DB : Set quiz.is_active = true
API -> WS : Broadcast quiz_started
WS -> S : Quiz available notification

note over S : Students join quiz

loop For each student
    S -> API : POST /api/v1/events (quiz_session_started)
    API -> DB : Create quiz_session record
    API -> WS : Update participant count
end

note over T, S : Quiz in progress

loop For each question
    T -> WS : Display question
    WS -> S : Question displayed

    par Student responses
        S -> API : POST /api/v1/events (quiz_answer_submitted)
        API -> DB : Store quiz_submission
        API -> A : Update real-time scores
        A -> WS : Broadcast answer statistics
        WS -> T : Live answer distribution
    end
end

note over T : Teacher ends quiz
T -> API : PUT /api/v1/admin/quizzes/{id}/complete
API -> DB : Complete all quiz_sessions
API -> A : Generate final analytics
A -> DB : Store quiz_analytics
API -> WS : Broadcast quiz_completed
WS -> S : Show results
WS -> T : Final analytics available

@enduml
```

### Cube.dev Style Query Processing

```plantuml
@startuml Generic_Query_Processing
title Generic Query Processing (Cube.dev Style)

participant "Analytics Client" as AC
participant "Query API" as API
participant "Query Builder" as QB
participant "Query Cache" as Cache
participant "PostgreSQL" as DB
participant "Query Optimizer" as QO

AC -> API : POST /api/v1/query\n(cube-style query)
note right of AC : {measures: ["events.count"],\ndimensions: ["users.role"]}

API -> API : Validate query structure
API -> QB : Parse query request

QB -> QB : Determine required tables & joins
QB -> QB : Build SELECT clause from measures
QB -> QB : Build FROM clause with JOINs
QB -> QB : Build WHERE clause from filters
QB -> QB : Build GROUP BY from dimensions

QB -> QO : Optimize query plan
QO -> QO : Check available indexes
QO -> QO : Rewrite for performance
QO -> QB : Optimized SQL query

QB -> Cache : Check query cache
alt Cache Hit
    Cache -> QB : Return cached results
else Cache Miss
    QB -> DB : Execute SQL query
    DB -> QB : Query results
    QB -> Cache : Cache results\n(TTL based on data freshness)
end

QB -> QB : Format results for cube response
QB -> API : Formatted data
API -> AC : 200 OK (query results)

note over AC : Client renders charts/dashboards

@enduml
```

---

## 📊 API Design

### Event Ingestion Endpoints

```http
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

### Report Generation Endpoints

#### Student Performance Report
```http
GET /api/v1/reports/student-performance?student_id={uuid}&date_from={date}&date_to={date}&include_details={boolean}
```

#### Classroom Engagement Report
```http
GET /api/v1/reports/classroom-engagement?classroom_id={uuid}&date_from={date}&date_to={date}
```

#### Content Effectiveness Report
```http
GET /api/v1/reports/content-effectiveness?school_id={uuid}&content_type={string}&date_from={date}&date_to={date}
```

### Generic Query API (Cube.dev Style)

```http
POST /api/v1/query
Content-Type: application/json

{
  "measures": ["events.count", "users.avg_session_duration"],
  "dimensions": ["users.role", "events.event_type"],
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
  ]
}
```

---

## 🚀 Quick Start Guide

### Prerequisites
- Go 1.21+
- PostgreSQL 13+
- Redis (optional, for caching)

### Setup Instructions

1. **Clone and Setup**
```bash
git clone <repository>
cd jio-supperr
go mod download
```

2. **Database Setup**
```bash
# Create database
createdb reporting_db

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=reporting_db
```

3. **Run Migrations and Seed Data**
```bash
# Run the demo (includes migrations and seeding)
go run cmd/demo/main.go
```

4. **Start the Server**
```bash
# Start the reporting server
go run cmd/reporting-server/main.go
```

5. **Access the API**
- Health Check: http://localhost:8080/health
- API Documentation: http://localhost:8080/docs
- Student Performance: http://localhost:8080/api/v1/reports/student-performance?student_id={uuid}

### Running the Demo

The demo script demonstrates all three required report types:

```bash
go run cmd/demo/main.go
```

This will:
- ✅ Create and populate the database with sample data (1,000 schools, 30,000 classrooms, 900,000 users)
- ✅ Generate **Student Performance Analysis** reports
- ✅ Generate **Classroom Engagement Metrics** reports
- ✅ Generate **Content Effectiveness Evaluation** reports
- ✅ Demonstrate **Cube.dev style generic queries**
- ✅ Save sample reports as JSON files for review

---

## 📈 Sample Report Outputs

### 1. Student Performance Analysis
```json
{
  "student_id": "user_123",
  "student_name": "John Doe",
  "period": {"from": "2024-01-01", "to": "2024-01-31"},
  "overall_stats": {
    "avg_quiz_score": 85.5,
    "completion_rate": 92.3,
    "engagement_score": 78.9,
    "performance_trend": "improving",
    "active_days": 28,
    "avg_daily_minutes": 45.2
  },
  "recommendations": [
    "Student is performing well. Continue current approach and consider advanced challenges."
  ]
}
```

### 2. Classroom Engagement Metrics
```json
{
  "classroom_id": "class_101",
  "classroom_name": "Mathematics Grade 5",
  "engagement_metrics": {
    "participation_rate": 87.5,
    "avg_session_duration": 45.2,
    "overall_engagement_score": 82.1,
    "collaboration_events": 234,
    "whiteboard_usage_minutes": 680,
    "notebook_usage_minutes": 920
  }
}
```

### 3. Content Effectiveness Evaluation
```json
{
  "content_analytics": {
    "total_content": 156,
    "avg_engagement_score": 73.2,
    "share_rate": 15.8,
    "interaction_rate": 68.4
  },
  "recommendations": [
    {
      "type": "create_more",
      "description": "Focus on creating more interactive content formats",
      "priority": "high"
    }
  ]
}
```

---

## 🔧 Architecture Features

### ✅ Scalability Features
- **Event-driven architecture** with message queues for async processing
- **Database connection pooling** and optimized indexes
- **Redis caching** for frequently accessed data
- **Materialized views** for complex analytical queries
- **Horizontal scaling** support for microservices

### ✅ Data Collection Strategy
- **Real-time event ingestion** from both mobile applications
- **Batch processing** for historical data analysis
- **Comprehensive event taxonomy** covering user interactions
- **Cross-application synchronization** tracking

### ✅ Reporting Capabilities
- **Student Performance Analytics** - Individual learning progression
- **Classroom Engagement Metrics** - Group collaboration analysis
- **Content Effectiveness Evaluation** - Content usage optimization
- **Real-time dashboards** with live session tracking
- **Export capabilities** (JSON, CSV, Excel)

### ✅ Generic Query Engine
- **Cube.dev compatible** query interface
- **Flexible measures and dimensions** for custom analytics
- **Time-series analysis** with various granularities
- **Filter and aggregation** support
- **Query optimization** and result caching

---

## 🎯 Deliverables Completed

✅ **1. Detailed Technical Design Document** - This README with comprehensive architecture

✅ **2. Database Schema Design** - Complete ERD with performance optimizations

✅ **3. API Design** - RESTful endpoints with authentication and authorization

✅ **4. Sequence Diagrams** - Key reporting workflows documented

✅ **5. Working Prototype** - Demonstrates data ingestion and storage

✅ **6. Three Report Types**:
   - Student Performance Analysis
   - Classroom Engagement Metrics
   - Content Effectiveness Evaluation

✅ **7. Bonus: Generic Query Engine** - Cube.dev style query capabilities

---

## 📝 Technologies Used

- **Backend**: Go 1.21, Gin Framework, GORM
- **Database**: PostgreSQL with time-series optimizations
- **Caching**: Redis for session and query caching
- **Documentation**: PlantUML for diagrams
- **API Design**: RESTful with JWT authentication
- **Architecture**: Clean Architecture with Domain-Driven Design

---

## 🔗 Additional Resources

- [Sequence Diagrams](SEQUENCE_DIAGRAMS.md) - Detailed workflow diagrams
- [Database Migrations](internal/seedmigrations/) - SQL schema definitions
- [API Handlers](internal/handlers/) - Request/response implementations
- [Sample Data](internal/seedutils/) - Test data generation utilities

---

*This reporting framework provides a scalable, comprehensive solution for educational analytics, designed to handle the specified scale of 1,000 schools with 900,000 users while providing actionable insights for educators and administrators.*