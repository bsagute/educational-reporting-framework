-- Educational Reporting Framework Schema
-- Migration 001: Core reporting tables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Schools table
CREATE TABLE schools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    district VARCHAR(255),
    region VARCHAR(100),
    contact_email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Classrooms table
CREATE TABLE classrooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    grade_level INTEGER,
    subject VARCHAR(100),
    teacher_id UUID,
    max_students INTEGER DEFAULT 30,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) NOT NULL CHECK (role IN ('teacher', 'student', 'admin')),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP
);

-- User classroom associations
CREATE TABLE user_classrooms (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    classroom_id UUID NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('teacher', 'student')),
    enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (user_id, classroom_id)
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    application VARCHAR(50) NOT NULL CHECK (application IN ('whiteboard', 'notebook')),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    device_info JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Content table
CREATE TABLE content (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    title VARCHAR(255),
    content_type VARCHAR(50) NOT NULL CHECK (content_type IN ('note', 'drawing', 'document', 'quiz', 'whiteboard_session')),
    content_data JSONB,
    file_size_bytes BIGINT DEFAULT 0,
    is_shared BOOLEAN DEFAULT FALSE,
    share_permissions JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quizzes table
CREATE TABLE quizzes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    classroom_id UUID NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    total_questions INTEGER DEFAULT 0,
    time_limit_minutes INTEGER,
    max_attempts INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT FALSE,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quiz questions table
CREATE TABLE quiz_questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) NOT NULL CHECK (question_type IN ('multiple_choice', 'true_false', 'short_answer', 'essay')),
    options JSONB,
    correct_answer TEXT,
    points INTEGER DEFAULT 1,
    order_index INTEGER NOT NULL,
    explanation TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quiz_id, order_index)
);

-- Events table (main event store)
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    school_id UUID REFERENCES schools(id) ON DELETE SET NULL,
    application VARCHAR(50) CHECK (application IN ('whiteboard', 'notebook')),
    timestamp TIMESTAMP NOT NULL,
    metadata JSONB,
    device_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quiz submissions table
CREATE TABLE quiz_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    submitted_answer TEXT,
    is_correct BOOLEAN,
    points_earned INTEGER DEFAULT 0,
    time_spent_seconds INTEGER,
    attempt_number INTEGER DEFAULT 1,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quiz_id, student_id, question_id, attempt_number)
);

-- Quiz sessions table (tracks overall quiz attempts)
CREATE TABLE quiz_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    total_score INTEGER DEFAULT 0,
    max_possible_score INTEGER DEFAULT 0,
    percentage_score DECIMAL(5,2),
    time_spent_seconds INTEGER,
    attempt_number INTEGER DEFAULT 1,
    is_completed BOOLEAN DEFAULT FALSE
);

-- Performance indexes
CREATE INDEX idx_events_timestamp ON events USING BTREE(timestamp);
CREATE INDEX idx_events_user_id ON events USING BTREE(user_id);
CREATE INDEX idx_events_classroom_id ON events USING BTREE(classroom_id);
CREATE INDEX idx_events_school_id ON events USING BTREE(school_id);
CREATE INDEX idx_events_type_timestamp ON events USING BTREE(event_type, timestamp);
CREATE INDEX idx_events_session_id ON events USING BTREE(session_id);

CREATE INDEX idx_sessions_user_time ON sessions USING BTREE(user_id, start_time);
CREATE INDEX idx_sessions_classroom_time ON sessions USING BTREE(classroom_id, start_time);
CREATE INDEX idx_sessions_application ON sessions USING BTREE(application);

CREATE INDEX idx_quiz_submissions_quiz_student ON quiz_submissions USING BTREE(quiz_id, student_id);
CREATE INDEX idx_quiz_submissions_student_time ON quiz_submissions USING BTREE(student_id, submitted_at);

CREATE INDEX idx_quiz_sessions_student ON quiz_sessions USING BTREE(student_id);
CREATE INDEX idx_quiz_sessions_quiz ON quiz_sessions USING BTREE(quiz_id);
CREATE INDEX idx_quiz_sessions_completed ON quiz_sessions USING BTREE(completed_at) WHERE completed_at IS NOT NULL;

CREATE INDEX idx_users_school ON users USING BTREE(school_id);
CREATE INDEX idx_users_role ON users USING BTREE(role);
CREATE INDEX idx_users_last_active ON users USING BTREE(last_active);

CREATE INDEX idx_classrooms_school ON classrooms USING BTREE(school_id);
CREATE INDEX idx_classrooms_teacher ON classrooms USING BTREE(teacher_id);

CREATE INDEX idx_content_creator ON content USING BTREE(creator_id);
CREATE INDEX idx_content_classroom ON content USING BTREE(classroom_id);
CREATE INDEX idx_content_type ON content USING BTREE(content_type);
CREATE INDEX idx_content_created ON content USING BTREE(created_at);

-- Partial indexes for active data
CREATE INDEX idx_active_sessions ON sessions USING BTREE(user_id, start_time) WHERE end_time IS NULL;
CREATE INDEX idx_active_quizzes ON quizzes USING BTREE(classroom_id, start_time) WHERE is_active = TRUE;
CREATE INDEX idx_incomplete_quiz_sessions ON quiz_sessions USING BTREE(student_id, started_at) WHERE is_completed = FALSE;

-- Add foreign key for teacher_id in classrooms (referencing users)
ALTER TABLE classrooms ADD CONSTRAINT fk_classrooms_teacher
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE SET NULL;

-- Add triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_schools_updated_at BEFORE UPDATE ON schools
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_classrooms_updated_at BEFORE UPDATE ON classrooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_content_updated_at BEFORE UPDATE ON content
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_quizzes_updated_at BEFORE UPDATE ON quizzes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();