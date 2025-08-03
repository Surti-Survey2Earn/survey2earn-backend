// internal/routes/routes.go
package routes

import (
	"survey2earn-backend/internal/handler"
	"survey2earn-backend/internal/middleware"
	"survey2earn-backend/internal/service"
	"survey2earn-backend/internal/repository"
	"survey2earn-backend/internal/database"
	"survey2earn-backend/internal/config"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, cfg *config.Config, db *database.Database) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	surveyRepo := repository.NewSurveyRepository(db.DB)
	responseRepo := repository.NewResponseRepository(db.DB)
	rewardRepo := repository.NewRewardRepository(db.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)
	surveyService := service.NewSurveyService(surveyRepo, userRepo, rewardRepo)
	responseService := service.NewResponseService(responseRepo, surveyRepo, rewardRepo, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	surveyHandler := handler.NewSurveyHandler(surveyService)
	responseHandler := handler.NewResponseHandler(responseService)

	// API version group
	api := router.Group("/api/" + cfg.Server.APIVersion)
	{
		// Public routes (no authentication required)
		public := api.Group("/")
		{
			// Authentication routes
			auth := public.Group("auth")
			{
				auth.POST("/login", authHandler.Login)
				auth.POST("/register", authHandler.Register)
				auth.POST("/refresh", authHandler.RefreshToken)
				auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
			}

			// Public survey routes
			public.GET("/surveys", surveyHandler.GetPublicSurveys)
			public.GET("/surveys/:id", surveyHandler.GetSurvey)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			user := protected.Group("user")
			{
				user.GET("/profile", authHandler.GetProfile)
				user.PUT("/profile", authHandler.UpdateProfile)
				user.GET("/stats", authHandler.GetUserStats)
			}

			// Survey management routes
			surveys := protected.Group("surveys")
			{
				surveys.POST("/", surveyHandler.CreateSurvey)
				surveys.GET("/my", surveyHandler.GetUserSurveys)
				surveys.PUT("/:id", surveyHandler.UpdateSurvey)
				surveys.DELETE("/:id", surveyHandler.DeleteSurvey)
				surveys.POST("/:id/publish", surveyHandler.PublishSurvey)
				// surveys.GET("/:id/analytics", surveyHandler.GetSurveyAnalytics) // Future implementation
			}

			// Survey response routes
			responses := protected.Group("responses")
			{
				responses.POST("/start", responseHandler.StartSurvey)
				responses.GET("/", responseHandler.GetUserResponses)
				responses.GET("/:id", responseHandler.GetResponse)
				responses.GET("/:id/progress", responseHandler.GetResponseProgress)
				responses.POST("/:id/answers", responseHandler.SubmitAnswers)
				responses.PUT("/:response_id/questions/:question_id", responseHandler.UpdateAnswer)
				responses.POST("/:id/abandon", responseHandler.AbandonSurvey)
				responses.POST("/complete", responseHandler.CompleteSurvey)
			}

			// Reward and transaction routes (future implementation)
			rewards := protected.Group("rewards")
			{
				rewards.GET("/balance", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get user balance - not implemented"})
				})
				rewards.GET("/transactions", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get transaction history - not implemented"})
				})
				rewards.POST("/withdraw", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Withdraw rewards - not implemented"})
				})
			}
		}

		// Admin routes (future implementation)
		admin := api.Group("admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/surveys", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin survey management - not implemented"})
			})
			admin.GET("/users", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin user management - not implemented"})
			})
			admin.GET("/analytics", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin analytics - not implemented"})
			})
		}
	}
}

// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"
	"survey2earn-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token (this would use your JWT service)
		// For now, we'll use a mock validation
		userID, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", userID)
		c.Next()
	})
}

// AdminMiddleware checks if user has admin privileges
func AdminMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User authentication required",
			})
			c.Abort()
			return
		}

		// Check if user is admin (mock implementation)
		isAdmin := checkAdminStatus(userID)
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// Mock token validation - replace with actual JWT validation
func validateToken(token string) (uint, error) {
	// This is a mock implementation
	// In a real application, you would:
	// 1. Parse the JWT token
	// 2. Validate the signature
	// 3. Check expiration
	// 4. Extract user ID from claims
	
	if token == "mock-valid-token" {
		return 1, nil // Return mock user ID
	}
	return 0, errors.New("invalid token")
}

// Mock admin status check - replace with actual implementation
func checkAdminStatus(userID uint) bool {
	// This is a mock implementation
	// In a real application, you would check the user's role in the database
	return userID == 1 // Mock: user ID 1 is admin
}

// internal/repository/interfaces.go
package repository

import (
	"survey2earn-backend/internal/models"
	"survey2earn-backend/internal/dto"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByWalletAddress(address string) (*models.User, error)
	Update(user *models.User) error
	UpdateBalance(userID uint, earned, xp float64) error
	GetStats(userID uint) (*models.UserStats, error)
}

type SurveyRepository interface {
	Create(survey *models.Survey) error
	Update(survey *models.Survey) error
	GetByID(id uint) (*models.Survey, error)
	GetByUserID(userID uint, status string, page, limit int) ([]models.Survey, int64, error)
	GetPublicSurveys(page, limit int, category, status string) ([]models.Survey, int64, error)
	Delete(id uint) error
	DeleteQuestions(surveyID uint) error
	PublishWithRewardPool(survey *models.Survey, pool *models.RewardPool) error
	UpdateStatistics(surveyID uint) error
}

type ResponseRepository interface {
	Create(response *models.Response) error
	Update(response *models.Response) error
	GetByID(id uint) (*models.Response, error)
	GetWithAnswers(id uint) (*models.Response, error)
	GetByUserID(userID uint, req *dto.ListResponsesRequest) ([]models.Response, int64, error)
	HasUserResponded(userID, surveyID uint) (bool, error)
	UpsertAnswer(answer *models.Answer) error
}

type RewardRepository interface {
	GetPoolBySurveyID(surveyID uint) (*models.RewardPool, error)
	ProcessReward(pool *models.RewardPool, transaction *models.RewardTransaction) error
	CreateTransaction(transaction *models.RewardTransaction) error
	UpdatePool(pool *models.RewardPool) error
}

// internal/repository/user_repository.go
package repository

import (
	"survey2earn-backend/internal/models"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByWalletAddress(address string) (*models.User, error) {
	var user models.User
	err := r.db.Where("wallet_address = ?", address).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdateBalance(userID uint, earned, xp float64) error {
	// Mock implementation
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"total_earned": gorm.Expr("total_earned + ?", earned),
		}).Error
}

func (r *userRepository) GetStats(userID uint) (*models.UserStats, error) {
	// Mock implementation
	return &models.UserStats{
		UserID: userID,
	}, nil
}

// internal/repository/survey_repository.go  
package repository

import (
	"survey2earn-backend/internal/models"
	"gorm.io/gorm"
)

type surveyRepository struct {
	db *gorm.DB
}

func NewSurveyRepository(db *gorm.DB) SurveyRepository {
	return &surveyRepository{db: db}
}

func (r *surveyRepository) Create(survey *models.Survey) error {
	return r.db.Create(survey).Error
}

func (r *surveyRepository) Update(survey *models.Survey) error {
	return r.db.Save(survey).Error
}

func (r *surveyRepository) GetByID(id uint) (*models.Survey, error) {
	var survey models.Survey
	err := r.db.Preload("Questions").Preload("Creator").First(&survey, id).Error
	return &survey, err
}

func (r *surveyRepository) GetByUserID(userID uint, status string, page, limit int) ([]models.Survey, int64, error) {
	var surveys []models.Survey
	var total int64

	query := r.db.Model(&models.Survey{}).Where("creator_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Creator").Offset(offset).Limit(limit).Find(&surveys).Error

	return surveys, total, err
}

func (r *surveyRepository) GetPublicSurveys(page, limit int, category, status string) ([]models.Survey, int64, error) {
	var surveys []models.Survey
	var total int64

	query := r.db.Model(&models.Survey{}).Where("is_public = ?", true)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Creator").Offset(offset).Limit(limit).Find(&surveys).Error

	return surveys, total, err
}

func (r *surveyRepository) Delete(id uint) error {
	return r.db.Delete(&models.Survey{}, id).Error
}

func (r *surveyRepository) DeleteQuestions(surveyID uint) error {
	return r.db.Where("survey_id = ?", surveyID).Delete(&models.Question{}).Error
}

func (r *surveyRepository) PublishWithRewardPool(survey *models.Survey, pool *models.RewardPool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(survey).Error; err != nil {
			return err
		}
		return tx.Create(pool).Error
	})
}

func (r *surveyRepository) UpdateStatistics(surveyID uint) error {
	// Mock implementation - update survey statistics
	return nil
}

// Add imports at the top of routes.go
import (
	"errors"
)