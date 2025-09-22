package api

import (
	"reporting-framework/internal/config"
	"reporting-framework/internal/handlers"
	"reporting-framework/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	config *config.Config
	router *gin.Engine
}

func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	server := &Server{
		db:     db,
		config: cfg,
		router: gin.Default(),
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Initialize handlers
	eventHandler := handlers.NewEventHandler(s.db)
	sessionHandler := handlers.NewSessionHandler(s.db)
	quizHandler := handlers.NewQuizHandler(s.db)
	reportHandler := handlers.NewReportHandler(s.db)
	analyticsHandler := handlers.NewAnalyticsHandler(s.db)
	crudHandler := handlers.NewCRUDHandler(s.db)

	// Middleware
	s.router.Use(middleware.CORS())
	s.router.Use(middleware.RequestLogger())

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Authentication middleware for protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(s.config))

		// Event tracking endpoints
		events := protected.Group("/events")
		{
			events.POST("/batch", eventHandler.BatchInsert)
		}

		// Session management endpoints
		sessions := protected.Group("/sessions")
		{
			sessions.POST("/start", sessionHandler.StartSession)
			sessions.POST("/:id/end", sessionHandler.EndSession)
			sessions.GET("/:id", sessionHandler.GetSession)
		}

		// Quiz management endpoints
		quizzes := protected.Group("/quizzes")
		{
			quizzes.POST("", quizHandler.CreateQuiz)
			quizzes.PUT("/:id", quizHandler.UpdateQuiz)
			quizzes.POST("/:id/responses", quizHandler.SubmitResponse)
			quizzes.GET("/:id", quizHandler.GetQuiz)
		}

		// Reporting endpoints
		reports := protected.Group("/reports")
		{
			reports.GET("/students/:id/performance", reportHandler.GetStudentPerformance)
			reports.GET("/classrooms/:id/engagement", reportHandler.GetClassroomEngagement)
			reports.GET("/classrooms/:id/performance-trends", reportHandler.GetClassroomTrends)
			reports.GET("/content/effectiveness", reportHandler.GetContentEffectiveness)
		}

		// Analytics endpoints (generic query framework)
		analytics := protected.Group("/analytics")
		{
			analytics.POST("/query", analyticsHandler.ExecuteQuery)
		}

		// WebSocket endpoint for real-time data
		live := v1.Group("/live")
		{
			live.GET("/classroom/:id", func(c *gin.Context) {
				handlers.HandleWebSocket(c, s.db)
			})
		}

		// CRUD endpoints for basic data management
		crud := v1.Group("/")
		{
			// School management
			crud.GET("/schools", crudHandler.GetSchools)
			crud.GET("/schools/:id", crudHandler.GetSchool)
			crud.POST("/schools", crudHandler.CreateSchool)

			// User management
			crud.GET("/users", crudHandler.GetUsers)
			crud.GET("/users/:id", crudHandler.GetUser)
			crud.POST("/users", crudHandler.CreateUser)

			// Classroom management
			crud.GET("/classrooms", crudHandler.GetClassrooms)
			crud.GET("/classrooms/:id", crudHandler.GetClassroom)
			crud.POST("/classrooms", crudHandler.CreateClassroom)
		}
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}