package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GenericQueryBuilder handles cube.dev style queries
type GenericQueryBuilder struct {
	db *gorm.DB
}

// NewGenericQueryBuilder creates a new query builder
func NewGenericQueryBuilder(db *gorm.DB) *GenericQueryBuilder {
	return &GenericQueryBuilder{db: db}
}


// CubeSchema defines the available measures and dimensions
type CubeSchema struct {
	Measures   map[string]MeasureDefinition   `json:"measures"`
	Dimensions map[string]DimensionDefinition `json:"dimensions"`
}

type MeasureDefinition struct {
	Type        string `json:"type"`        // count, sum, avg, min, max
	SQL         string `json:"sql"`         // SQL expression
	Table       string `json:"table"`       // Source table
	Description string `json:"description"` // Human readable description
}

type DimensionDefinition struct {
	Type        string `json:"type"`        // string, number, time
	SQL         string `json:"sql"`         // SQL expression
	Table       string `json:"table"`       // Source table
	Description string `json:"description"` // Human readable description
}

// GetSchema returns the available cube schema
func (q *GenericQueryBuilder) GetSchema() CubeSchema {
	return CubeSchema{
		Measures: map[string]MeasureDefinition{
			// Event measures
			"events.count": {
				Type:        "count",
				SQL:         "COUNT(*)",
				Table:       "events",
				Description: "Total number of events",
			},
			"events.unique_users": {
				Type:        "count",
				SQL:         "COUNT(DISTINCT user_id)",
				Table:       "events",
				Description: "Number of unique users generating events",
			},

			// Session measures
			"sessions.count": {
				Type:        "count",
				SQL:         "COUNT(*)",
				Table:       "sessions",
				Description: "Total number of sessions",
			},
			"sessions.avg_duration": {
				Type:        "avg",
				SQL:         "AVG(duration_seconds / 60.0)",
				Table:       "sessions",
				Description: "Average session duration in minutes",
			},
			"sessions.total_duration": {
				Type:        "sum",
				SQL:         "SUM(duration_seconds / 60.0)",
				Table:       "sessions",
				Description: "Total session duration in minutes",
			},

			// User measures
			"users.count": {
				Type:        "count",
				SQL:         "COUNT(*)",
				Table:       "users",
				Description: "Total number of users",
			},
			"users.active_count": {
				Type:        "count",
				SQL:         "COUNT(*)",
				Table:       "users",
				Description: "Number of active users (with recent activity)",
			},

			// Quiz measures
			"quizzes.count": {
				Type:        "count",
				SQL:         "COUNT(*)",
				Table:       "quizzes",
				Description: "Total number of quizzes",
			},
			"quiz_sessions.avg_score": {
				Type:        "avg",
				SQL:         "AVG(percentage_score)",
				Table:       "quiz_sessions",
				Description: "Average quiz score percentage",
			},
			"quiz_sessions.completion_rate": {
				Type:        "avg",
				SQL:         "AVG(CASE WHEN is_completed THEN 1.0 ELSE 0.0 END) * 100",
				Table:       "quiz_sessions",
				Description: "Quiz completion rate percentage",
			},

			// Content measures
			"content.count": {
				Type:        "count",
				SQL:         "COUNT(*)",
				Table:       "content",
				Description: "Total number of content items",
			},
			"content.avg_file_size": {
				Type:        "avg",
				SQL:         "AVG(file_size_bytes / 1024.0 / 1024.0)",
				Table:       "content",
				Description: "Average content file size in MB",
			},
		},
		Dimensions: map[string]DimensionDefinition{
			// Time dimensions
			"time.date": {
				Type:        "time",
				SQL:         "DATE(created_at)",
				Table:       "events",
				Description: "Date of the event",
			},
			"time.hour": {
				Type:        "time",
				SQL:         "DATE_TRUNC('hour', created_at)",
				Table:       "events",
				Description: "Hour of the event",
			},
			"time.week": {
				Type:        "time",
				SQL:         "DATE_TRUNC('week', created_at)",
				Table:       "events",
				Description: "Week of the event",
			},
			"time.month": {
				Type:        "time",
				SQL:         "DATE_TRUNC('month', created_at)",
				Table:       "events",
				Description: "Month of the event",
			},

			// User dimensions
			"users.role": {
				Type:        "string",
				SQL:         "role",
				Table:       "users",
				Description: "User role (teacher, student, admin)",
			},
			"users.school_id": {
				Type:        "string",
				SQL:         "school_id::text",
				Table:       "users",
				Description: "School identifier",
			},

			// Event dimensions
			"events.type": {
				Type:        "string",
				SQL:         "event_type",
				Table:       "events",
				Description: "Type of event",
			},
			"events.application": {
				Type:        "string",
				SQL:         "application",
				Table:       "events",
				Description: "Application source (whiteboard, notebook)",
			},

			// Session dimensions
			"sessions.application": {
				Type:        "string",
				SQL:         "application",
				Table:       "sessions",
				Description: "Session application type",
			},

			// Content dimensions
			"content.type": {
				Type:        "string",
				SQL:         "content_type",
				Table:       "content",
				Description: "Type of content",
			},

			// School/Classroom dimensions
			"schools.name": {
				Type:        "string",
				SQL:         "s.name",
				Table:       "schools",
				Description: "School name",
			},
			"classrooms.name": {
				Type:        "string",
				SQL:         "c.name",
				Table:       "classrooms",
				Description: "Classroom name",
			},
			"classrooms.grade_level": {
				Type:        "number",
				SQL:         "c.grade_level",
				Table:       "classrooms",
				Description: "Classroom grade level",
			},
			"classrooms.subject": {
				Type:        "string",
				SQL:         "c.subject",
				Table:       "classrooms",
				Description: "Classroom subject",
			},
		},
	}
}

// ExecuteQuery executes a cube.dev style query
func (q *GenericQueryBuilder) ExecuteQuery(queryReq interface{}) ([]map[string]interface{}, error) {
	// Cast to proper type
	req, ok := queryReq.(struct {
		Measures       []string `json:"measures"`
		Dimensions     []string `json:"dimensions"`
		TimeDimensions []struct {
			Dimension   string   `json:"dimension"`
			Granularity string   `json:"granularity"`
			DateRange   []string `json:"dateRange"`
		} `json:"timeDimensions"`
		Filters []struct {
			Member   string   `json:"member"`
			Operator string   `json:"operator"`
			Values   []string `json:"values"`
		} `json:"filters"`
		Order [][]string `json:"order"`
		Limit int        `json:"limit"`
	})
	if !ok {
		return nil, fmt.Errorf("invalid query request type")
	}

	schema := q.GetSchema()

	// Build SQL query
	query, err := q.buildSQL(req, schema)
	if err != nil {
		return nil, err
	}

	// Execute query
	var results []map[string]interface{}
	if err := q.db.Raw(query).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return results, nil
}

// buildSQL constructs the SQL query from the cube request
func (q *GenericQueryBuilder) buildSQL(req struct {
	Measures       []string `json:"measures"`
	Dimensions     []string `json:"dimensions"`
	TimeDimensions []struct {
		Dimension   string   `json:"dimension"`
		Granularity string   `json:"granularity"`
		DateRange   []string `json:"dateRange"`
	} `json:"timeDimensions"`
	Filters []struct {
		Member   string   `json:"member"`
		Operator string   `json:"operator"`
		Values   []string `json:"values"`
	} `json:"filters"`
	Order [][]string `json:"order"`
	Limit int        `json:"limit"`
}, schema CubeSchema) (string, error) {

	// Determine primary table and joins needed
	tables := q.determineTables(req.Measures, req.Dimensions, schema)
	primaryTable := q.determinePrimaryTable(tables)

	// Build SELECT clause
	selectClauses := []string{}

	// Add measures
	for _, measure := range req.Measures {
		if def, exists := schema.Measures[measure]; exists {
			alias := strings.ReplaceAll(measure, ".", "_")
			selectClauses = append(selectClauses, fmt.Sprintf("%s AS %s", def.SQL, alias))
		}
	}

	// Add dimensions
	for _, dimension := range req.Dimensions {
		if def, exists := schema.Dimensions[dimension]; exists {
			alias := strings.ReplaceAll(dimension, ".", "_")
			selectClauses = append(selectClauses, fmt.Sprintf("%s AS %s", def.SQL, alias))
		}
	}

	// Add time dimensions
	for _, timeDim := range req.TimeDimensions {
		if def, exists := schema.Dimensions[timeDim.Dimension]; exists {
			var sql string
			switch timeDim.Granularity {
			case "day":
				sql = fmt.Sprintf("DATE(%s)", def.SQL)
			case "week":
				sql = fmt.Sprintf("DATE_TRUNC('week', %s)", def.SQL)
			case "month":
				sql = fmt.Sprintf("DATE_TRUNC('month', %s)", def.SQL)
			case "hour":
				sql = fmt.Sprintf("DATE_TRUNC('hour', %s)", def.SQL)
			default:
				sql = def.SQL
			}
			alias := fmt.Sprintf("%s_%s", strings.ReplaceAll(timeDim.Dimension, ".", "_"), timeDim.Granularity)
			selectClauses = append(selectClauses, fmt.Sprintf("%s AS %s", sql, alias))
		}
	}

	if len(selectClauses) == 0 {
		return "", fmt.Errorf("no valid measures or dimensions specified")
	}

	// Build FROM clause with JOINs
	fromClause := q.buildFromClause(primaryTable, tables)

	// Build WHERE clause
	whereClause := q.buildWhereClause(req.Filters, req.TimeDimensions, schema)

	// Build GROUP BY clause
	groupByClause := q.buildGroupByClause(req.Dimensions, req.TimeDimensions, schema)

	// Build ORDER BY clause
	orderByClause := q.buildOrderByClause(req.Order)

	// Build LIMIT clause
	limitClause := ""
	if req.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", req.Limit)
	}

	// Combine all parts
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectClauses, ", "), fromClause)

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	if groupByClause != "" {
		query += " GROUP BY " + groupByClause
	}

	if orderByClause != "" {
		query += " ORDER BY " + orderByClause
	}

	if limitClause != "" {
		query += " " + limitClause
	}

	return query, nil
}

// Helper methods for SQL building

func (q *GenericQueryBuilder) determineTables(measures, dimensions []string, schema CubeSchema) map[string]bool {
	tables := make(map[string]bool)

	for _, measure := range measures {
		if def, exists := schema.Measures[measure]; exists {
			tables[def.Table] = true
		}
	}

	for _, dimension := range dimensions {
		if def, exists := schema.Dimensions[dimension]; exists {
			tables[def.Table] = true
		}
	}

	return tables
}

func (q *GenericQueryBuilder) determinePrimaryTable(tables map[string]bool) string {
	// Priority order for primary table selection
	priority := []string{"events", "sessions", "users", "quizzes", "content", "schools", "classrooms"}

	for _, table := range priority {
		if tables[table] {
			return table
		}
	}

	// Default to events if no priority table found
	return "events"
}

func (q *GenericQueryBuilder) buildFromClause(primaryTable string, tables map[string]bool) string {
	from := primaryTable

	// Add table aliases
	aliases := map[string]string{
		"events":     "e",
		"sessions":   "s",
		"users":      "u",
		"quizzes":    "q",
		"content":    "c",
		"schools":    "sch",
		"classrooms": "cl",
	}

	if alias, exists := aliases[primaryTable]; exists {
		from = fmt.Sprintf("%s %s", primaryTable, alias)
	}

	// Add necessary JOINs based on relationships
	joins := []string{}

	switch primaryTable {
	case "events":
		if tables["users"] {
			joins = append(joins, "LEFT JOIN users u ON e.user_id = u.id")
		}
		if tables["sessions"] {
			joins = append(joins, "LEFT JOIN sessions s ON e.session_id = s.id")
		}
		if tables["classrooms"] {
			joins = append(joins, "LEFT JOIN classrooms cl ON e.classroom_id = cl.id")
		}
		if tables["schools"] {
			joins = append(joins, "LEFT JOIN schools sch ON e.school_id = sch.id")
		}

	case "sessions":
		if tables["users"] {
			joins = append(joins, "LEFT JOIN users u ON s.user_id = u.id")
		}
		if tables["classrooms"] {
			joins = append(joins, "LEFT JOIN classrooms cl ON s.classroom_id = cl.id")
		}
		if tables["schools"] && !tables["classrooms"] {
			joins = append(joins, "LEFT JOIN classrooms cl ON s.classroom_id = cl.id")
			joins = append(joins, "LEFT JOIN schools sch ON cl.school_id = sch.id")
		} else if tables["schools"] {
			joins = append(joins, "LEFT JOIN schools sch ON cl.school_id = sch.id")
		}

	case "users":
		if tables["schools"] {
			joins = append(joins, "LEFT JOIN schools sch ON u.school_id = sch.id")
		}
		if tables["classrooms"] {
			joins = append(joins, "LEFT JOIN user_classrooms uc ON u.id = uc.user_id")
			joins = append(joins, "LEFT JOIN classrooms cl ON uc.classroom_id = cl.id")
		}
	}

	if len(joins) > 0 {
		from += " " + strings.Join(joins, " ")
	}

	return from
}

func (q *GenericQueryBuilder) buildWhereClause(filters []struct {
	Member   string   `json:"member"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}, timeDimensions []struct {
	Dimension   string   `json:"dimension"`
	Granularity string   `json:"granularity"`
	DateRange   []string `json:"dateRange"`
}, schema CubeSchema) string {

	conditions := []string{}

	// Add filter conditions
	for _, filter := range filters {
		if def, exists := schema.Dimensions[filter.Member]; exists {
			condition := q.buildFilterCondition(def.SQL, filter.Operator, filter.Values)
			if condition != "" {
				conditions = append(conditions, condition)
			}
		}
	}

	// Add time range conditions
	for _, timeDim := range timeDimensions {
		if def, exists := schema.Dimensions[timeDim.Dimension]; exists && len(timeDim.DateRange) == 2 {
			condition := fmt.Sprintf("%s BETWEEN '%s' AND '%s'", def.SQL, timeDim.DateRange[0], timeDim.DateRange[1])
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " AND ")
}

func (q *GenericQueryBuilder) buildFilterCondition(sql, operator string, values []string) string {
	switch operator {
	case "equals":
		if len(values) == 1 {
			return fmt.Sprintf("%s = '%s'", sql, values[0])
		}
	case "in":
		if len(values) > 0 {
			quotedValues := make([]string, len(values))
			for i, v := range values {
				quotedValues[i] = fmt.Sprintf("'%s'", v)
			}
			return fmt.Sprintf("%s IN (%s)", sql, strings.Join(quotedValues, ", "))
		}
	case "gt":
		if len(values) == 1 {
			return fmt.Sprintf("%s > '%s'", sql, values[0])
		}
	case "gte":
		if len(values) == 1 {
			return fmt.Sprintf("%s >= '%s'", sql, values[0])
		}
	case "lt":
		if len(values) == 1 {
			return fmt.Sprintf("%s < '%s'", sql, values[0])
		}
	case "lte":
		if len(values) == 1 {
			return fmt.Sprintf("%s <= '%s'", sql, values[0])
		}
	case "contains":
		if len(values) == 1 {
			return fmt.Sprintf("%s ILIKE '%%%s%%'", sql, values[0])
		}
	}
	return ""
}

func (q *GenericQueryBuilder) buildGroupByClause(dimensions []string, timeDimensions []struct {
	Dimension   string   `json:"dimension"`
	Granularity string   `json:"granularity"`
	DateRange   []string `json:"dateRange"`
}, schema CubeSchema) string {

	groupByClauses := []string{}

	// Add dimension group bys
	for _, dimension := range dimensions {
		if def, exists := schema.Dimensions[dimension]; exists {
			groupByClauses = append(groupByClauses, def.SQL)
		}
	}

	// Add time dimension group bys
	for _, timeDim := range timeDimensions {
		if def, exists := schema.Dimensions[timeDim.Dimension]; exists {
			var sql string
			switch timeDim.Granularity {
			case "day":
				sql = fmt.Sprintf("DATE(%s)", def.SQL)
			case "week":
				sql = fmt.Sprintf("DATE_TRUNC('week', %s)", def.SQL)
			case "month":
				sql = fmt.Sprintf("DATE_TRUNC('month', %s)", def.SQL)
			case "hour":
				sql = fmt.Sprintf("DATE_TRUNC('hour', %s)", def.SQL)
			default:
				sql = def.SQL
			}
			groupByClauses = append(groupByClauses, sql)
		}
	}

	if len(groupByClauses) == 0 {
		return ""
	}

	return strings.Join(groupByClauses, ", ")
}

func (q *GenericQueryBuilder) buildOrderByClause(order [][]string) string {
	if len(order) == 0 {
		return ""
	}

	orderClauses := []string{}
	for _, orderItem := range order {
		if len(orderItem) >= 2 {
			column := strings.ReplaceAll(orderItem[0], ".", "_")
			direction := strings.ToUpper(orderItem[1])
			if direction == "ASC" || direction == "DESC" {
				orderClauses = append(orderClauses, fmt.Sprintf("%s %s", column, direction))
			}
		}
	}

	return strings.Join(orderClauses, ", ")
}

// GetAvailableMetrics returns all available measures and dimensions
func (q *GenericQueryBuilder) GetAvailableMetrics() gin.H {
	schema := q.GetSchema()

	measures := make([]gin.H, 0, len(schema.Measures))
	for name, def := range schema.Measures {
		measures = append(measures, gin.H{
			"name":        name,
			"type":        def.Type,
			"description": def.Description,
			"table":       def.Table,
		})
	}

	dimensions := make([]gin.H, 0, len(schema.Dimensions))
	for name, def := range schema.Dimensions {
		dimensions = append(dimensions, gin.H{
			"name":        name,
			"type":        def.Type,
			"description": def.Description,
			"table":       def.Table,
		})
	}

	return gin.H{
		"measures":   measures,
		"dimensions": dimensions,
	}
}