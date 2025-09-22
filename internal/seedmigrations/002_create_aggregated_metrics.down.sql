-- Drop materialized view and function
DROP FUNCTION IF EXISTS refresh_classroom_performance_mv();
DROP MATERIALIZED VIEW IF EXISTS mv_classroom_performance;

-- Drop tables in reverse order
DROP TABLE IF EXISTS hourly_event_aggregates CASCADE;
DROP TABLE IF EXISTS active_sessions CASCADE;
DROP TABLE IF EXISTS quiz_analytics CASCADE;
DROP TABLE IF EXISTS content_metrics CASCADE;
DROP TABLE IF EXISTS weekly_school_metrics CASCADE;
DROP TABLE IF EXISTS daily_classroom_metrics CASCADE;
DROP TABLE IF EXISTS daily_user_metrics CASCADE;