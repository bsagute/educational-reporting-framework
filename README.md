# Educational Applications Reporting Framework

A comprehensive reporting and analytics framework designed for educational applications, specifically optimized for Whiteboard and Notebook app ecosystems with real-time quiz synchronization capabilities.

## 🎯 Overview

This framework captures, stores, and analyzes user interactions and performance metrics across educational applications, providing detailed insights into student performance, classroom engagement, and content effectiveness.

**Scale**: Designed to handle data from ~1,000 schools, 30,000 classrooms, and 900,000 students.

## 🏗️ Architecture

### System Components

- **API Gateway**: RESTful API with JWT/API key authentication
- **WebSocket Service**: Real-time data streaming for live dashboards
- **Analytics Engine**: Generic query framework inspired by Cube.dev
- **Database Layer**: PostgreSQL with optimized schemas and indexing
- **Cache Layer**: Redis for session management and report caching

### Technology Stack

**Backend**
- Go 1.21+ (Gin framework)
- PostgreSQL 14+ with TimescaleDB extension
- Redis 6+ for caching
- GORM for database operations
- WebSocket support via Gorilla WebSocket

**Infrastructure**
- Docker containers
- Environment-based configuration
- Structured logging
- Comprehensive error handling

## 📊 Key Features

### 1. Real-time Data Collection
- Event-based tracking methodology
- Batch processing for high-throughput scenarios
- Comprehensive session management
- Multi-application support (Whiteboard/Notebook)

### 2. Analytics & Reporting
- **Student Performance Analysis**: Quiz scores, engagement metrics, learning trends
- **Classroom Engagement Metrics**: Real-time participation, response times, activity levels
- **Content Effectiveness Evaluation**: Quiz performance, completion rates, time analysis
- **Generic Query Framework**: Cube.dev-inspired measures and dimensions API

### 3. Real-time Capabilities
- Live classroom dashboards via WebSocket
- Real-time quiz participation tracking
- Instant engagement score calculations
- Live event streaming

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd jio-super
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Set up database**
   ```bash
   # Create PostgreSQL database
   createdb reporting_db

   # Run migrations and seed data
   go run cmd/server/main.go  # This will run migrations
   go run cmd/seed/main.go    # This will seed sample data
   ```

5. **Start the server**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

### Configuration

Key environment variables:

```env
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/reporting_db?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
WHITEBOARD_API_KEY=wb_key_123
NOTEBOOK_API_KEY=nb_key_456
```

## 📖 API Documentation

### Authentication

The API supports two authentication methods:

1. **JWT Bearer Token**
   ```bash
   Authorization: Bearer <jwt_token>
   ```

2. **API Key** (for applications)
   ```bash
   X-API-Key: <api_key>
   ```

### Core Endpoints

#### Event Tracking
```http
POST /api/v1/events/batch
Content-Type: application/json

{
  "events": [
    {
      "event_type": "quiz_answer_submitted",
      "timestamp": "2024-01-15T10:30:00Z",
      "user_id": "user_123",
      "session_id": "session_456",
      "application": "notebook",
      "payload": {
        "quiz_id": "quiz_001",
        "answer": "A",
        "time_taken_seconds": 45
      }
    }
  ]
}
```

#### Student Performance
```http
GET /api/v1/reports/students/{student_id}/performance?start_date=2024-01-01&end_date=2024-01-31
```

#### Classroom Engagement
```http
GET /api/v1/reports/classrooms/{classroom_id}/engagement?date=2024-01-15
```

#### Generic Analytics Query
```http
POST /api/v1/analytics/query
{
  "measures": [
    "quiz_responses.average_score",
    "sessions.total_duration"
  ],
  "dimensions": [
    "users.role",
    "classrooms.subject"
  ],
  "filters": [
    {
      "dimension": "time.date",
      "operator": "gte",
      "value": "2024-01-01"
    }
  ]
}
```

#### Real-time WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/live/classroom/{classroom_id}');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Live update:', data);
};
```

## 📈 Data Models

### Core Entities

- **Schools**: Educational institutions
- **Users**: Teachers, students, administrators
- **Classrooms**: Learning spaces with subject/grade metadata
- **Sessions**: App usage sessions with duration tracking
- **Events**: All user interactions and system events
- **Quizzes**: Quiz content with questions and metadata
- **Quiz Responses**: Student answers with scoring

### Event Structure
```json
{
  "event_id": "uuid",
  "event_type": "quiz_answer_submitted",
  "user_id": "uuid",
  "session_id": "uuid",
  "timestamp": "ISO8601",
  "application": "notebook|whiteboard",
  "payload": {
    "quiz_id": "uuid",
    "question_id": "uuid",
    "answer": "A",
    "time_taken_seconds": 45
  },
  "metadata": {
    "device_type": "tablet",
    "app_version": "2.1.0"
  }
}
```

## 🔍 Analytics Framework

### Predefined Measures
- `quiz_responses.average_score`: Average quiz performance
- `quiz_responses.completion_rate`: Quiz completion percentage
- `sessions.total_duration`: Total session time in minutes
- `sessions.average_duration`: Average session duration
- `events.count`: Total number of events
- `users.active_count`: Count of active users

### Predefined Dimensions
- `users.role`: User role (teacher/student/admin)
- `classrooms.subject`: Subject being taught
- `schools.name`: School name
- `time.date`: Date dimension
- `applications.type`: Application type

### Time Granularities
- hour, day, week, month, quarter, year

## 🗄️ Database Schema

The framework uses a normalized PostgreSQL schema with the following key tables:

- `schools` - Educational institutions
- `users` - All system users
- `classrooms` - Learning spaces
- `enrollments` - Student-classroom relationships
- `sessions` - App usage sessions
- `events` - All tracked interactions
- `quizzes` - Quiz definitions
- `quiz_questions` - Individual questions
- `quiz_responses` - Student answers
- `daily_user_stats` - Aggregated daily metrics
- `classroom_analytics` - Classroom-level aggregations

Performance optimizations include:
- Strategic indexing on timestamp, user_id, session_id
- JSONB columns for flexible event payloads
- Partitioning strategies for large event tables

## 📊 Sample Queries

### Student Performance Over Time
```sql
SELECT
  DATE(qr.submitted_at) as date,
  AVG(CASE WHEN qr.points_earned IS NOT NULL THEN
    (qr.points_earned / qq.points) * 100 END) as avg_score
FROM quiz_responses qr
JOIN quiz_questions qq ON qr.question_id = qq.id
WHERE qr.student_id = $1
GROUP BY DATE(qr.submitted_at)
ORDER BY date;
```

### Classroom Engagement Analysis
```sql
SELECT
  c.subject,
  COUNT(DISTINCT s.user_id) as active_students,
  AVG(s.duration_seconds) / 60 as avg_session_minutes
FROM sessions s
JOIN classrooms c ON s.classroom_id = c.id
WHERE s.start_time >= NOW() - INTERVAL '7 days'
GROUP BY c.subject;
```

## 🔧 Development

### Project Structure
```
├── cmd/
│   ├── server/          # Main server application
│   └── seed/           # Database seeding utility
├── internal/
│   ├── api/            # HTTP server setup
│   ├── config/         # Configuration management
│   ├── database/       # Database connection & migrations
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # HTTP middleware
│   └── models/         # Database models
├── TECHNICAL_DESIGN.md # Detailed technical documentation
├── go.mod              # Go module definition
└── README.md           # This file
```

### Running Tests
```bash
go test ./...
```

### Seeding Sample Data
```bash
go run cmd/seed/main.go
```

This creates:
- 3 sample schools
- 15 teachers
- 30 classrooms
- 900 students
- 150 quizzes with questions
- 1000+ sessions with events
- Thousands of quiz responses

## 🚦 Performance Considerations

### Scalability Features
- **Database Optimizations**: Strategic indexing, connection pooling
- **Caching Strategy**: Multi-layer caching (Redis + in-memory)
- **Batch Processing**: Efficient bulk operations for high-throughput scenarios
- **Real-time Processing**: WebSocket-based live updates
- **Partitioning**: Time-based partitioning for large tables

### Load Testing
The system is designed to handle:
- 10,000+ concurrent users
- 1M+ events per hour
- Sub-second query response times
- Real-time updates for 1000+ concurrent dashboards

## 🛠️ Deployment

### Docker Setup
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Environment Variables
See `.env.example` for all available configuration options.

## 📝 Technical Design Document

For detailed technical specifications, architecture decisions, and implementation details, see [TECHNICAL_DESIGN.md](./TECHNICAL_DESIGN.md).

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For questions or support:
- Create an issue in the repository
- Review the technical design document
- Check the API documentation

---

**Built with 💡 for educational excellence**