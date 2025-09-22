-- Educational Reporting Framework Schema
-- Migration 002: Aggregated metrics and performance tables

-- Daily user metrics for performance reporting
CREATE TABLE daily_user_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    session_count INTEGER DEFAULT 0,
    total_session_duration_seconds INTEGER DEFAULT 0,
    avg_session_duration_seconds DECIMAL(10,2) DEFAULT 0,
    events_count INTEGER DEFAULT 0,
    quiz_attempts INTEGER DEFAULT 0,
    quiz_completions INTEGER DEFAULT 0,
    avg_quiz_score DECIMAL(5,2),
    total_quiz_score INTEGER DEFAULT 0,
    max_quiz_score INTEGER DEFAULT 0,
    content_created_count INTEGER DEFAULT 0,
    content_viewed_count INTEGER DEFAULT 0,
    whiteboard_events INTEGER DEFAULT 0,
    notebook_events INTEGER DEFAULT 0,
    collaboration_events INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

-- Daily classroom metrics for engagement reporting
CREATE TABLE daily_classroom_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    classroom_id UUID NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_students INTEGER DEFAULT 0,
    active_students_count INTEGER DEFAULT 0,
    participation_rate DECIMAL(5,2) DEFAULT 0,
    total_sessions INTEGER DEFAULT 0,
    avg_session_duration_minutes DECIMAL(10,2) DEFAULT 0,
    total_quiz_sessions INTEGER DEFAULT 0,
    avg_quiz_completion_rate DECIMAL(5,2) DEFAULT 0,
    avg_class_quiz_score DECIMAL(5,2),
    content_created_count INTEGER DEFAULT 0,
    content_shared_count INTEGER DEFAULT 0,
    whiteboard_usage_minutes INTEGER DEFAULT 0,
    notebook_usage_minutes INTEGER DEFAULT 0,
    sync_events_count INTEGER DEFAULT 0,
    engagement_score DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(classroom_id, date)
);

-- Weekly school metrics for administrative reporting
CREATE TABLE weekly_school_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    week_start_date DATE NOT NULL,
    total_classrooms INTEGER DEFAULT 0,
    active_classrooms INTEGER DEFAULT 0,
    total_users INTEGER DEFAULT 0,
    active_users INTEGER DEFAULT 0,
    total_students INTEGER DEFAULT 0,
    active_students INTEGER DEFAULT 0,
    total_teachers INTEGER DEFAULT 0,
    active_teachers INTEGER DEFAULT 0,
    total_sessions INTEGER DEFAULT 0,
    avg_daily_sessions DECIMAL(10,2) DEFAULT 0,
    total_quiz_sessions INTEGER DEFAULT 0,
    avg_school_engagement DECIMAL(5,2) DEFAULT 0,
    total_content_created INTEGER DEFAULT 0,
    platform_adoption_rate DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(school_id, week_start_date)
);

-- Content effectiveness metrics
CREATE TABLE content_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    school_id UUID REFERENCES schools(id) ON DELETE SET NULL,
    content_type VARCHAR(50) NOT NULL,
    view_count INTEGER DEFAULT 0,
    unique_viewers INTEGER DEFAULT 0,
    avg_view_duration_seconds DECIMAL(10,2) DEFAULT 0,
    interaction_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    effectiveness_score DECIMAL(5,2) DEFAULT 0,
    last_viewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(content_id)
);

-- Quiz performance analytics
CREATE TABLE quiz_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    classroom_id UUID NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
    total_attempts INTEGER DEFAULT 0,
    unique_participants INTEGER DEFAULT 0,
    completion_rate DECIMAL(5,2) DEFAULT 0,
    avg_score DECIMAL(5,2) DEFAULT 0,
    avg_time_spent_minutes DECIMAL(10,2) DEFAULT 0,
    difficulty_score DECIMAL(5,2) DEFAULT 0,
    question_analytics JSONB,
    last_attempt_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quiz_id)
);

-- Real-time session tracking for live analytics
CREATE TABLE active_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    application VARCHAR(50) NOT NULL,
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_activity VARCHAR(100),
    activity_metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id)
);

-- Event aggregation for real-time dashboards
CREATE TABLE hourly_event_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hour_timestamp TIMESTAMP NOT NULL,
    school_id UUID REFERENCES schools(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    application VARCHAR(50),
    event_count INTEGER DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(hour_timestamp, school_id, classroom_id, event_type, application)
);

-- Indexes for aggregated metrics tables
CREATE INDEX idx_daily_user_metrics_date ON daily_user_metrics USING BTREE(date);
CREATE INDEX idx_daily_user_metrics_user_date ON daily_user_metrics USING BTREE(user_id, date);
CREATE INDEX idx_daily_user_metrics_school_date ON daily_user_metrics USING BTREE(school_id, date);

CREATE INDEX idx_daily_classroom_metrics_date ON daily_classroom_metrics USING BTREE(date);
CREATE INDEX idx_daily_classroom_metrics_classroom_date ON daily_classroom_metrics USING BTREE(classroom_id, date);
CREATE INDEX idx_daily_classroom_metrics_school_date ON daily_classroom_metrics USING BTREE(school_id, date);

CREATE INDEX idx_weekly_school_metrics_week ON weekly_school_metrics USING BTREE(week_start_date);
CREATE INDEX idx_weekly_school_metrics_school_week ON weekly_school_metrics USING BTREE(school_id, week_start_date);

CREATE INDEX idx_content_metrics_content ON content_metrics USING BTREE(content_id);
CREATE INDEX idx_content_metrics_classroom ON content_metrics USING BTREE(classroom_id);
CREATE INDEX idx_content_metrics_type ON content_metrics USING BTREE(content_type);
CREATE INDEX idx_content_metrics_effectiveness ON content_metrics USING BTREE(effectiveness_score DESC);

CREATE INDEX idx_quiz_analytics_quiz ON quiz_analytics USING BTREE(quiz_id);
CREATE INDEX idx_quiz_analytics_classroom ON quiz_analytics USING BTREE(classroom_id);
CREATE INDEX idx_quiz_analytics_completion ON quiz_analytics USING BTREE(completion_rate DESC);

CREATE INDEX idx_active_sessions_session ON active_sessions USING BTREE(session_id);
CREATE INDEX idx_active_sessions_user ON active_sessions USING BTREE(user_id);
CREATE INDEX idx_active_sessions_classroom ON active_sessions USING BTREE(classroom_id);
CREATE INDEX idx_active_sessions_heartbeat ON active_sessions USING BTREE(last_heartbeat);

CREATE INDEX idx_hourly_aggregates_timestamp ON hourly_event_aggregates USING BTREE(hour_timestamp);
CREATE INDEX idx_hourly_aggregates_school_time ON hourly_event_aggregates USING BTREE(school_id, hour_timestamp);
CREATE INDEX idx_hourly_aggregates_classroom_time ON hourly_event_aggregates USING BTREE(classroom_id, hour_timestamp);

-- Add updated_at triggers for new tables
CREATE TRIGGER update_daily_user_metrics_updated_at BEFORE UPDATE ON daily_user_metrics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_daily_classroom_metrics_updated_at BEFORE UPDATE ON daily_classroom_metrics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_weekly_school_metrics_updated_at BEFORE UPDATE ON weekly_school_metrics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_content_metrics_updated_at BEFORE UPDATE ON content_metrics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_quiz_analytics_updated_at BEFORE UPDATE ON quiz_analytics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create materialized views for performance
CREATE MATERIALIZED VIEW mv_classroom_performance AS
SELECT
    c.id as classroom_id,
    c.name as classroom_name,
    c.school_id,
    s.name as school_name,
    COUNT(DISTINCT uc.user_id) as total_students,
    AVG(dum.avg_quiz_score) as avg_class_quiz_score,
    AVG(dum.total_session_duration_seconds / 60.0) as avg_daily_minutes,
    AVG(dcm.participation_rate) as avg_participation_rate,
    AVG(dcm.engagement_score) as avg_engagement_score
FROM classrooms c
JOIN schools s ON c.school_id = s.id
LEFT JOIN user_classrooms uc ON c.id = uc.classroom_id AND uc.is_active = TRUE
LEFT JOIN daily_user_metrics dum ON uc.user_id = dum.user_id
    AND dum.date >= CURRENT_DATE - INTERVAL '30 days'
LEFT JOIN daily_classroom_metrics dcm ON c.id = dcm.classroom_id
    AND dcm.date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY c.id, c.name, c.school_id, s.name;

CREATE UNIQUE INDEX idx_mv_classroom_performance_classroom ON mv_classroom_performance(classroom_id);

-- Create refresh function for materialized view
CREATE OR REPLACE FUNCTION refresh_classroom_performance_mv()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_classroom_performance;
END;
$$ LANGUAGE plpgsql;