// internal/dto/auth.go
package dto

import "time"

// LoginRequest represents the login request
type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Message       string `json:"message" binding:"required"`
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// TokenResponse represents token refresh response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// UserProfileResponse represents user profile information
type UserProfileResponse struct {
	ID              uint     `json:"id"`
	WalletAddress   string   `json:"wallet_address"`
	Username        *string  `json:"username"`
	Email           *string  `json:"email"`
	Bio             *string  `json:"bio"`
	ProfilePicture  *string  `json:"profile_picture"`
	ReputationScore float64  `json:"reputation_score"`
	TotalEarned     float64  `json:"total_earned"`
	TotalResponses  int      `json:"total_responses"`
	TotalSurveys    int      `json:"total_surveys"`
	IsActive        bool     `json:"is_active"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Username       *string `json:"username"`
	Email          *string `json:"email"`
	Bio            *string `json:"bio"`
	ProfilePicture *string `json:"profile_picture"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	UserID               uint       `json:"user_id"`
	TotalSurveysCreated  int        `json:"total_surveys_created"`
	TotalSurveysAnswered int        `json:"total_surveys_answered"`
	TotalEarned          float64    `json:"total_earned"`
	TotalSpent           float64    `json:"total_spent"`
	AverageRating        float64    `json:"average_rating"`
	LastActivityAt       *time.Time `json:"last_activity_at"`
}

// Additional missing DTOs for survey analytics
type SurveyAnalyticsResponse struct {
	SurveyID           uint                     `json:"survey_id"`
	TotalResponses     int                      `json:"total_responses"`
	CompletionRate     float64                  `json:"completion_rate"`
	AverageRating      float64                  `json:"average_rating"`
	AverageDuration    int                      `json:"average_duration"`
	Demographics       DemographicsData         `json:"demographics"`
	QuestionAnalytics  []QuestionAnalytics      `json:"question_analytics"`
	ResponseTrends     []ResponseTrendData      `json:"response_trends"`
}

type DemographicsData struct {
	AgeGroups      map[string]int `json:"age_groups"`
	Countries      map[string]int `json:"countries"`
	Languages      map[string]int `json:"languages"`
}

type QuestionAnalytics struct {
	QuestionID       uint                   `json:"question_id"`
	QuestionText     string                 `json:"question_text"`
	QuestionType     string                 `json:"question_type"`
	ResponseCount    int                    `json:"response_count"`
	SkipRate         float64                `json:"skip_rate"`
	AverageTimeSpent int                    `json:"average_time_spent"`
	AnswerDistribution map[string]interface{} `json:"answer_distribution"`
}

type ResponseTrendData struct {
	Date      string `json:"date"`
	Count     int    `json:"count"`
	Completed int    `json:"completed"`
}