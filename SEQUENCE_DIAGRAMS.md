# Educational Reporting Framework - Sequence Diagrams

## 1. Real-time Event Ingestion Workflow

```mermaid
sequenceDiagram
    participant WA as Whiteboard App
    participant NA as Notebook App
    participant RF as Reporting Framework
    participant DB as PostgreSQL
    participant Redis as Redis Cache
    participant MQ as Message Queue

    Note over WA, NA: User interactions generate events

    WA->>RF: POST /api/v1/events (batch events)
    Note right of WA: Events: whiteboard_draw, content_created, etc.

    RF->>RF: Validate events & authenticate user
    RF->>RF: Enrich events with context (school_id, session_id)

    par Parallel Processing
        RF->>DB: Batch insert events
        RF->>Redis: Cache latest session state
        RF->>MQ: Queue aggregation jobs
    end

    RF-->>WA: 201 Created (event_ids)

    Note over NA: Student submits quiz answer
    NA->>RF: POST /api/v1/events (quiz_answer_submitted)
    RF->>RF: Validate quiz submission
    RF->>DB: Store quiz submission
    RF->>RF: Calculate real-time quiz scores
    RF->>Redis: Update live quiz analytics
    RF-->>NA: 201 Created

    Note over MQ: Background aggregation
    MQ->>RF: Process aggregation job
    RF->>DB: Update daily_user_metrics
    RF->>DB: Update daily_classroom_metrics
```

## 2. Student Performance Report Generation

```mermaid
sequenceDiagram
    participant Teacher as Teacher Dashboard
    participant API as Reporting API
    participant Cache as Redis Cache
    participant DB as PostgreSQL
    participant Analytics as Analytics Engine

    Teacher->>API: GET /api/v1/reports/student-performance?student_id=xyz&date_from=2024-01-01
    API->>API: Authenticate & authorize teacher access

    API->>Cache: Check cached report
    alt Cache Hit
        Cache-->>API: Return cached data
        API-->>Teacher: 200 OK (cached report)
    else Cache Miss
        API->>DB: Query daily_user_metrics
        API->>DB: Query quiz_sessions for details
        API->>DB: Query events for engagement data

        API->>Analytics: Calculate engagement score
        Analytics->>Analytics: Process learning progression
        Analytics-->>API: Calculated metrics

        API->>API: Format response
        API->>Cache: Cache report (TTL: 1 hour)
        API-->>Teacher: 200 OK (performance report)
    end

    Note over Teacher: Teacher views detailed analytics
    Teacher->>API: GET /api/v1/reports/student-performance?include_details=true
    API->>DB: Query detailed quiz performance
    API->>DB: Query learning progression timeline
    API-->>Teacher: 200 OK (detailed report)
```

## 3. Real-time Quiz Session Workflow

```mermaid
sequenceDiagram
    participant Teacher as Teacher (Whiteboard App)
    participant Students as Students (Notebook App)
    participant API as Reporting API
    participant DB as PostgreSQL
    participant WS as WebSocket Server
    participant Analytics as Real-time Analytics

    Note over Teacher: Teacher creates and starts quiz
    Teacher->>API: POST /api/v1/admin/quizzes (create quiz)
    API->>DB: Store quiz and questions
    API-->>Teacher: 201 Created (quiz_id)

    Teacher->>API: PUT /api/v1/admin/quizzes/{id}/activate
    API->>DB: Set quiz.is_active = true
    API->>WS: Broadcast quiz_started event
    WS-->>Students: Quiz available notification

    Note over Students: Students join quiz session

    loop For each student
        Students->>API: POST /api/v1/events (quiz_session_started)
        API->>DB: Create quiz_session record
        API->>WS: Update live participant count
    end

    Note over Teacher, Students: Quiz in progress

    loop For each question
        Teacher->>WS: Display question (via whiteboard)
        WS-->>Students: Question displayed

        par Student responses
            Students->>API: POST /api/v1/events (quiz_answer_submitted)
            API->>DB: Store quiz_submission
            API->>Analytics: Update real-time scores
            Analytics->>WS: Broadcast answer statistics
            WS-->>Teacher: Live answer distribution
        end
    end

    Note over Teacher: Teacher ends quiz
    Teacher->>API: PUT /api/v1/admin/quizzes/{id}/complete
    API->>DB: Complete all active quiz_sessions
    API->>Analytics: Generate final quiz analytics
    Analytics->>DB: Store quiz_analytics
    API->>WS: Broadcast quiz_completed
    WS-->>Students: Quiz completed, show results
    WS-->>Teacher: Final analytics available
```

## 4. Classroom Engagement Analytics Workflow

```mermaid
sequenceDiagram
    participant Admin as School Admin
    participant API as Reporting API
    participant DB as PostgreSQL
    participant MV as Materialized Views
    participant Export as Export Service

    Admin->>API: GET /api/v1/reports/classroom-engagement?classroom_id=abc&period=30d
    API->>API: Authenticate admin access

    par Data Retrieval
        API->>DB: Query daily_classroom_metrics
        API->>DB: Query user engagement data
        API->>MV: Query mv_classroom_performance
    end

    API->>API: Calculate participation rates
    API->>API: Generate engagement trends
    API->>API: Identify top/bottom performers

    API-->>Admin: 200 OK (engagement report)

    Note over Admin: Admin requests detailed export
    Admin->>API: POST /api/v1/reports/export (classroom engagement)
    API->>Export: Queue export job
    Export->>DB: Generate comprehensive dataset
    Export->>Export: Create CSV/Excel file
    Export->>API: Export completed
    API-->>Admin: 200 OK (download_url)
```

## 5. Generic Query Processing (Cube.dev style)

```mermaid
sequenceDiagram
    participant Client as Analytics Client
    participant API as Query API
    participant QB as Query Builder
    participant Cache as Query Cache
    participant DB as PostgreSQL
    participant Optimizer as Query Optimizer

    Client->>API: POST /api/v1/query (cube-style query)
    Note right of Client: {measures: ["events.count"], dimensions: ["users.role"]}

    API->>API: Validate query structure
    API->>QB: Parse query request

    QB->>QB: Determine required tables & joins
    QB->>QB: Build SELECT clause from measures
    QB->>QB: Build FROM clause with JOINs
    QB->>QB: Build WHERE clause from filters
    QB->>QB: Build GROUP BY from dimensions

    QB->>Optimizer: Optimize query plan
    Optimizer->>Optimizer: Check for available indexes
    Optimizer->>Optimizer: Rewrite for performance
    Optimizer-->>QB: Optimized SQL query

    QB->>Cache: Check query cache
    alt Cache Hit
        Cache-->>QB: Return cached results
    else Cache Miss
        QB->>DB: Execute SQL query
        DB-->>QB: Query results
        QB->>Cache: Cache results (TTL based on data freshness)
    end

    QB->>QB: Format results for cube response
    QB-->>API: Formatted data
    API-->>Client: 200 OK (query results)

    Note over Client: Client renders charts/dashboards
```

## 6. Data Aggregation and Metrics Update Workflow

```mermaid
sequenceDiagram
    participant Scheduler as Cron Scheduler
    participant Worker as Background Worker
    participant DB as PostgreSQL
    participant MV as Materialized Views
    participant Alerts as Alert Service

    Note over Scheduler: Daily aggregation job
    Scheduler->>Worker: Trigger daily metrics aggregation

    Worker->>DB: Lock aggregation process

    par Parallel Aggregation
        Worker->>DB: Aggregate daily_user_metrics
        Worker->>DB: Aggregate daily_classroom_metrics
        Worker->>DB: Aggregate weekly_school_metrics
    end

    Worker->>DB: Update content_metrics
    Worker->>DB: Update quiz_analytics

    Worker->>MV: Refresh materialized views
    Note right of MV: mv_classroom_performance, etc.

    Worker->>Worker: Calculate metric deltas
    Worker->>Alerts: Check alert thresholds

    alt Alert Conditions Met
        Alerts->>Alerts: Generate alert notifications
        Alerts->>Worker: Send notifications to admins
    end

    Worker->>DB: Update aggregation metadata
    Worker->>DB: Release aggregation lock

    Note over Worker: Aggregation completed successfully
```

## 7. Content Effectiveness Analysis Workflow

```mermaid
sequenceDiagram
    participant Apps as Mobile Apps
    participant API as Reporting API
    participant ML as ML Pipeline
    participant DB as PostgreSQL
    participant Insights as Insights Engine

    Note over Apps: Users interact with content

    loop Content Interactions
        Apps->>API: POST /api/v1/events (content_viewed, content_modified)
        API->>DB: Store interaction events
    end

    Note over ML: Scheduled content analysis
    ML->>DB: Extract content interaction patterns
    ML->>ML: Analyze view duration patterns
    ML->>ML: Calculate engagement scores
    ML->>ML: Identify effective content characteristics

    ML->>Insights: Generate content recommendations
    Insights->>Insights: Rank content by effectiveness
    Insights->>Insights: Identify improvement opportunities

    Insights->>DB: Store content_metrics
    Insights->>DB: Update effectiveness scores

    Note over API: Teacher requests content report
    API->>DB: Query content effectiveness data
    API->>Insights: Get content recommendations
    API-->>Apps: Content effectiveness report with recommendations
```

These sequence diagrams illustrate the key workflows in the educational reporting framework, showing how different components interact to provide real-time analytics, comprehensive reporting, and actionable insights for educational applications.