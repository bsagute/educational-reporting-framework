package handlers

import (
	"net/http"

	"reporting-framework/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CRUDHandler struct {
	db *gorm.DB
}

func NewCRUDHandler(db *gorm.DB) *CRUDHandler {
	return &CRUDHandler{db: db}
}

// School CRUD operations
func (h *CRUDHandler) GetSchools(c *gin.Context) {
	var schools []models.School
	if err := h.db.Find(&schools).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve schools",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, schools)
}

func (h *CRUDHandler) GetSchool(c *gin.Context) {
	id := c.Param("id")
	schoolID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid school ID format",
			},
		})
		return
	}

	var school models.School
	if err := h.db.First(&school, schoolID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "School not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve school",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, school)
}

func (h *CRUDHandler) CreateSchool(c *gin.Context) {
	var school models.School
	if err := c.ShouldBindJSON(&school); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	if err := h.db.Create(&school).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create school",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, school)
}

// User CRUD operations
func (h *CRUDHandler) GetUsers(c *gin.Context) {
	var users []models.User
	role := c.Query("role")

	query := h.db.Preload("School")
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve users",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *CRUDHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	var user models.User
	if err := h.db.Preload("School").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve user",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *CRUDHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create user",
				"details": err.Error(),
			},
		})
		return
	}

	// Reload with school data
	h.db.Preload("School").First(&user, user.ID)

	c.JSON(http.StatusCreated, user)
}

// Classroom CRUD operations
func (h *CRUDHandler) GetClassrooms(c *gin.Context) {
	var classrooms []models.Classroom
	schoolID := c.Query("school_id")

	query := h.db.Preload("School").Preload("Teacher")
	if schoolID != "" {
		if id, err := uuid.Parse(schoolID); err == nil {
			query = query.Where("school_id = ?", id)
		}
	}

	if err := query.Find(&classrooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve classrooms",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, classrooms)
}

func (h *CRUDHandler) GetClassroom(c *gin.Context) {
	id := c.Param("id")
	classroomID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid classroom ID format",
			},
		})
		return
	}

	var classroom models.Classroom
	if err := h.db.Preload("School").Preload("Teacher").First(&classroom, classroomID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "Classroom not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to retrieve classroom",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, classroom)
}

func (h *CRUDHandler) CreateClassroom(c *gin.Context) {
	var classroom models.Classroom
	if err := c.ShouldBindJSON(&classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	if err := h.db.Create(&classroom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create classroom",
				"details": err.Error(),
			},
		})
		return
	}

	// Reload with relationship data
	h.db.Preload("School").Preload("Teacher").First(&classroom, classroom.ID)

	c.JSON(http.StatusCreated, classroom)
}