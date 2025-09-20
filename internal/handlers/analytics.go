package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	db *gorm.DB
}

type QueryRequest struct {
	Measures      []string          `json:"measures" binding:"required"`
	Dimensions    []string          `json:"dimensions"`
	Filters       []Filter          `json:"filters"`
	TimeDimension *TimeDimension    `json:"time_dimension"`
	Limit         *int              `json:"limit"`
	Offset        *int              `json:"offset"`
}

type Filter struct {
	Dimension string      `json:"dimension" binding:"required"`
	Operator  string      `json:"operator" binding:"required"`
	Value     interface{} `json:"value" binding:"required"`
}

type TimeDimension struct {
	Dimension   string `json:"dimension" binding:"required"`
	Granularity string `json:"granularity" binding:"required"`
}

type QueryResult struct {
	Data []map[string]interface{} `json:"data"`
	Meta QueryMeta                `json:"meta"`
}

type QueryMeta struct {
	TotalRows int    `json:"total_rows"`
	QueryTime string `json:"query_time"`
}

// Predefined measures and dimensions mapping
var measureMap = map[string]string{
	"quiz_responses.average_score":     "AVG(CASE WHEN qr.points_earned IS NOT NULL THEN (qr.points_earned / qq.points) * 100 END)",
	"quiz_responses.completion_rate":   "COUNT(DISTINCT qr.student_id) * 100.0 / COUNT(DISTINCT e.user_id)",
	"quiz_responses.total_responses":   "COUNT(qr.id)",
	"sessions.total_duration":          "SUM(s.duration_seconds) / 60.0",
	"sessions.average_duration":        "AVG(s.duration_seconds) / 60.0",
	"sessions.count":                   "COUNT(s.id)",
	"events.count":                     "COUNT(e.id)",
	"users.active_count":               "COUNT(DISTINCT u.id)",
	"classrooms.engagement_score":      "AVG(ca.average_engagement_score)",
}

var dimensionMap = map[string]string{
	"users.role":           "u.role",
	"classrooms.subject":   "c.subject",
	"schools.name":         "sch.name",
	"time.date":            "DATE(s.start_time)",
	"time.hour":            "EXTRACT(HOUR FROM s.start_time)",
	"time.day_of_week":     "EXTRACT(DOW FROM s.start_time)",
	"applications.type":    "s.application",
	"quiz.difficulty":      "q.status",
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) ExecuteQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid query format",
				"details": err.Error(),
			},
		})
		return
	}

	startTime := time.Now()

	// Build the SQL query
	query, err := h.buildQuery(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "QUERY_BUILD_ERROR",
				"message": "Failed to build query",
				"details": err.Error(),
			},
		})
		return
	}

	// Execute the query
	var results []map[string]interface{}
	if err := h.db.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "QUERY_EXECUTION_ERROR",
				"message": "Failed to execute query",
				"details": err.Error(),
			},
		})
		return
	}

	queryTime := time.Since(startTime)

	response := QueryResult{
		Data: results,
		Meta: QueryMeta{
			TotalRows: len(results),
			QueryTime: queryTime.String(),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *AnalyticsHandler) buildQuery(req QueryRequest) (string, error) {
	// Build SELECT clause
	var selects []string

	// Add measures
	for _, measure := range req.Measures {
		if sqlExpr, exists := measureMap[measure]; exists {
			selects = append(selects, fmt.Sprintf("%s as \"%s\"", sqlExpr, measure))
		} else {
			return "", fmt.Errorf("unknown measure: %s", measure)
		}
	}

	// Add dimensions
	for _, dimension := range req.Dimensions {
		if sqlExpr, exists := dimensionMap[dimension]; exists {
			selects = append(selects, fmt.Sprintf("%s as \"%s\"", sqlExpr, dimension))
		} else {
			return "", fmt.Errorf("unknown dimension: %s", dimension)
		}
	}

	// Add time dimension if specified
	if req.TimeDimension != nil {
		timeExpr, err := h.buildTimeDimension(*req.TimeDimension)
		if err != nil {
			return "", err
		}
		selects = append(selects, fmt.Sprintf("%s as \"time\"", timeExpr))
	}

	selectClause := strings.Join(selects, ", ")

	// Build FROM clause with JOINs
	fromClause := h.buildFromClause(req)

	// Build WHERE clause
	whereClause, err := h.buildWhereClause(req.Filters)
	if err != nil {
		return "", err
	}

	// Build GROUP BY clause
	groupByClause := h.buildGroupByClause(req.Dimensions, req.TimeDimension)

	// Build ORDER BY clause
	orderByClause := ""
	if req.TimeDimension != nil {
		orderByClause = "ORDER BY time ASC"
	}

	// Build LIMIT and OFFSET
	limitClause := ""
	if req.Limit != nil {
		limitClause = fmt.Sprintf("LIMIT %d", *req.Limit)
		if req.Offset != nil {
			limitClause += fmt.Sprintf(" OFFSET %d", *req.Offset)
		}
	}

	// Combine all parts
	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, fromClause)

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	if groupByClause != "" {
		query += " GROUP BY " + groupByClause
	}

	if orderByClause != "" {
		query += " " + orderByClause
	}

	if limitClause != "" {
		query += " " + limitClause
	}

	return query, nil
}

func (h *AnalyticsHandler) buildFromClause(req QueryRequest) string {
	// Determine which tables we need based on measures and dimensions
	needsQuizResponses := false
	needsQuizQuestions := false
	needsQuizzes := false
	needsSessions := false
	needsEvents := false
	needsUsers := false
	needsClassrooms := false
	needsSchools := false
	needsClassroomAnalytics := false

	// Check measures
	for _, measure := range req.Measures {
		switch {
		case strings.HasPrefix(measure, "quiz_responses."):
			needsQuizResponses = true
			needsQuizQuestions = true
		case strings.HasPrefix(measure, "sessions."):
			needsSessions = true
		case strings.HasPrefix(measure, "events."):
			needsEvents = true
		case strings.HasPrefix(measure, "users."):
			needsUsers = true
		case strings.HasPrefix(measure, "classrooms."):
			needsClassroomAnalytics = true
			needsClassrooms = true
		}
	}

	// Check dimensions
	for _, dimension := range req.Dimensions {
		switch {
		case strings.HasPrefix(dimension, "users."):
			needsUsers = true
		case strings.HasPrefix(dimension, "classrooms."):
			needsClassrooms = true
		case strings.HasPrefix(dimension, "schools."):
			needsSchools = true
		case strings.HasPrefix(dimension, "time.") || strings.HasPrefix(dimension, "applications."):
			needsSessions = true
		}
	}

	// Start with the main table (most commonly used)
	var baseTable string
	var joins []string

	if needsSessions {
		baseTable = "sessions s"
		if needsUsers {
			joins = append(joins, "LEFT JOIN users u ON s.user_id = u.id")
		}
		if needsClassrooms {
			joins = append(joins, "LEFT JOIN classrooms c ON s.classroom_id = c.id")
		}
		if needsSchools {
			joins = append(joins, "LEFT JOIN schools sch ON c.school_id = sch.id")
		}
		if needsEvents {
			joins = append(joins, "LEFT JOIN events e ON s.id = e.session_id")
		}
	} else if needsQuizResponses {
		baseTable = "quiz_responses qr"
		if needsQuizQuestions {
			joins = append(joins, "JOIN quiz_questions qq ON qr.question_id = qq.id")
		}
		if needsQuizzes {
			joins = append(joins, "JOIN quizzes q ON qr.quiz_id = q.id")
		}
		if needsUsers {
			joins = append(joins, "LEFT JOIN users u ON qr.student_id = u.id")
		}
		if needsClassrooms {
			joins = append(joins, "LEFT JOIN classrooms c ON q.classroom_id = c.id")
		}
	} else if needsEvents {
		baseTable = "events e"
		if needsUsers {
			joins = append(joins, "LEFT JOIN users u ON e.user_id = u.id")
		}
		if needsClassrooms {
			joins = append(joins, "LEFT JOIN classrooms c ON e.classroom_id = c.id")
		}
	} else if needsClassroomAnalytics {
		baseTable = "classroom_analytics ca"
		if needsClassrooms {
			joins = append(joins, "JOIN classrooms c ON ca.classroom_id = c.id")
		}
	} else {
		baseTable = "users u"
		if needsClassrooms {
			joins = append(joins, "LEFT JOIN enrollments en ON u.id = en.user_id")
			joins = append(joins, "LEFT JOIN classrooms c ON en.classroom_id = c.id")
		}
	}

	fromClause := baseTable
	if len(joins) > 0 {
		fromClause += " " + strings.Join(joins, " ")
	}

	return fromClause
}

func (h *AnalyticsHandler) buildWhereClause(filters []Filter) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	for _, filter := range filters {
		condition, err := h.buildFilterCondition(filter)
		if err != nil {
			return "", err
		}
		conditions = append(conditions, condition)
	}

	return strings.Join(conditions, " AND "), nil
}

func (h *AnalyticsHandler) buildFilterCondition(filter Filter) (string, error) {
	dimension, exists := dimensionMap[filter.Dimension]
	if !exists {
		return "", fmt.Errorf("unknown filter dimension: %s", filter.Dimension)
	}

	switch filter.Operator {
	case "eq", "=":
		return fmt.Sprintf("%s = '%v'", dimension, filter.Value), nil
	case "ne", "!=":
		return fmt.Sprintf("%s != '%v'", dimension, filter.Value), nil
	case "gt", ">":
		return fmt.Sprintf("%s > '%v'", dimension, filter.Value), nil
	case "gte", ">=":
		return fmt.Sprintf("%s >= '%v'", dimension, filter.Value), nil
	case "lt", "<":
		return fmt.Sprintf("%s < '%v'", dimension, filter.Value), nil
	case "lte", "<=":
		return fmt.Sprintf("%s <= '%v'", dimension, filter.Value), nil
	case "in":
		if values, ok := filter.Value.([]interface{}); ok {
			var strValues []string
			for _, v := range values {
				strValues = append(strValues, fmt.Sprintf("'%v'", v))
			}
			return fmt.Sprintf("%s IN (%s)", dimension, strings.Join(strValues, ", ")), nil
		}
		return "", fmt.Errorf("'in' operator requires array value")
	case "contains":
		return fmt.Sprintf("%s ILIKE '%%%v%%'", dimension, filter.Value), nil
	default:
		return "", fmt.Errorf("unsupported operator: %s", filter.Operator)
	}
}

func (h *AnalyticsHandler) buildGroupByClause(dimensions []string, timeDimension *TimeDimension) string {
	var groups []string

	for _, dimension := range dimensions {
		if sqlExpr, exists := dimensionMap[dimension]; exists {
			groups = append(groups, sqlExpr)
		}
	}

	if timeDimension != nil {
		timeExpr, _ := h.buildTimeDimension(*timeDimension)
		groups = append(groups, timeExpr)
	}

	if len(groups) == 0 {
		return ""
	}

	return strings.Join(groups, ", ")
}

func (h *AnalyticsHandler) buildTimeDimension(td TimeDimension) (string, error) {
	baseDimension := dimensionMap[td.Dimension]
	if baseDimension == "" {
		return "", fmt.Errorf("unknown time dimension: %s", td.Dimension)
	}

	switch td.Granularity {
	case "hour":
		return fmt.Sprintf("DATE_TRUNC('hour', %s)", baseDimension), nil
	case "day":
		return fmt.Sprintf("DATE_TRUNC('day', %s)", baseDimension), nil
	case "week":
		return fmt.Sprintf("DATE_TRUNC('week', %s)", baseDimension), nil
	case "month":
		return fmt.Sprintf("DATE_TRUNC('month', %s)", baseDimension), nil
	case "quarter":
		return fmt.Sprintf("DATE_TRUNC('quarter', %s)", baseDimension), nil
	case "year":
		return fmt.Sprintf("DATE_TRUNC('year', %s)", baseDimension), nil
	default:
		return "", fmt.Errorf("unsupported time granularity: %s", td.Granularity)
	}
}