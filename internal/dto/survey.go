// internal/dto/survey.go
package dto

import (
	"time"
)

// CreateSurveyRequest represents the request to create a new survey
type CreateSurveyRequest struct {
	Title             string                   `json:"title" binding:"required,min=3,max=255"`
	Description       string                   `json:"description" binding:"required"`
	Category          string                   `json:"category" binding:"required"`
	EstimatedTime     string                   `json:"estimatedTime" binding:"required"`
	RewardAmount      float64                  `json:"rewardAmount" binding:"required,gt=0"`
	MaxParticipants   int                      `json:"maxParticipants" binding:"required,gt=0"`
	XpReward          int                      `json:"xpReward" binding:"required,gt=0"`
	Questions         []CreateQuestionRequest  `json:"questions" binding:"required,min=1"`
	IsAnonymous       bool                     `json:"isAnonymous"`
	IsPublic          bool                     `json:"isPublic"`
	RequireLogin      bool                     `json:"requireLogin"`
	AllowMultiple     bool                     `json:"allowMultiple"`
	StartDate         *time.Time               `json:"startDate"`
	EndDate           *time.Time               `json:"endDate"`
}

// CreateQuestionRequest represents a question in the survey creation request
type CreateQuestionRequest struct {
	Type        string                    `json:"type" binding:"required"`
	Title       string                    `json:"title" binding:"required"`
	Description string                    `json:"description"`
	Required    bool                      `json:"required"`
	Options     []QuestionOptionRequest   `json:"options"`
	MinLength   *int                      `json:"minLength"`
	MaxLength   *int                      `json:"maxLength"`
	MinValue    *float64                  `json:"minValue"`
	MaxValue    *float64                  `json:"maxValue"`
	Order       int                       `json:"order"`
}

// QuestionOptionRequest represents question option
type QuestionOptionRequest struct {
	ID    string `json:"id"`
	Label string `json:"label" binding:"required"`
	Value string `json:"value" binding:"required"`
	Order int    `json:"order"`
}

// UpdateSurveyRequest for updating draft surveys
type UpdateSurveyRequest struct {
	Title           *string                   `json:"title"`
	Description     *string                   `json:"description"`
	Category        *string                   `json:"category"`
	EstimatedTime   *string                   `json:"estimatedTime"`
	RewardAmount    *float64                  `json:"rewardAmount"`
	MaxParticipants *int                      `json:"maxParticipants"`
	XpReward        *int                      `json:"xpReward"`
	Questions       []CreateQuestionRequest   `json:"questions"`
	IsAnonymous     *bool                     `json:"isAnonymous"`
	IsPublic        *bool                     `json:"isPublic"`
	RequireLogin    *bool                     `json:"requireLogin"`
	AllowMultiple   *bool                     `json:"allowMultiple"`
}

// PublishSurveyRequest for publishing a survey
type PublishSurveyRequest struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

// SurveyResponse represents the survey response
type SurveyResponse struct {
	ID                uint                     `json:"id"`
	CreatorID         uint                     `json:"creator_id"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description"`
	Category          string                   `json:"category"`
	Status            string                   `json:"status"`
	MaxResponses      int                      `json:"max_responses"`
	RewardPerResponse float64                  `json:"reward_per_response"`
	TotalRewardPool   float64                  `json:"total_reward_pool"`
	EstimatedDuration int                      `json:"estimated_duration"`
	ResponseCount     int                      `json:"response_count"`
	CompletionRate    float64                  `json:"completion_rate"`
	AverageRating     float64                  `json:"average_rating"`
	IsAnonymous       bool                     `json:"is_anonymous"`
	IsPublic          bool                     `json:"is_public"`
	RequireLogin      bool                     `json:"require_login"`
	AllowMultiple     bool                     `json:"allow_multiple"`
	StartDate         *time.Time               `json:"start_date"`
	EndDate           *time.Time               `json:"end_date"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	Questions         []QuestionResponse       `json:"questions"`
	Creator           UserResponse             `json:"creator"`
}

// QuestionResponse represents question in response
type QuestionResponse struct {
	ID          uint                       `json:"id"`
	Type        string                     `json:"type"`
	Text        string                     `json:"text"`
	Description string                     `json:"description"`
	Required    bool                       `json:"required"`
	Order       int                        `json:"order"`
	Options     []QuestionOptionResponse   `json:"options"`
	MinLength   *int                       `json:"min_length"`
	MaxLength   *int                       `json:"max_length"`
	MinValue    *float64                   `json:"min_value"`
	MaxValue    *float64                   `json:"max_value"`
}

// QuestionOptionResponse represents question option in response
type QuestionOptionResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
	Order int    `json:"order"`
}

// UserResponse represents user in response
type UserResponse struct {
	ID              uint    `json:"id"`
	WalletAddress   string  `json:"wallet_address"`
	Username        *string `json:"username"`
	ReputationScore float64 `json:"reputation_score"`
}

// SurveyListResponse for listing surveys
type SurveyListResponse struct {
	Surveys    []SurveyItemResponse `json:"surveys"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

// SurveyItemResponse for survey list item
type SurveyItemResponse struct {
	ID                uint         `json:"id"`
	Title             string       `json:"title"`
	Description       string       `json:"description"`
	Category          string       `json:"category"`
	Status            string       `json:"status"`
	RewardPerResponse float64      `json:"reward_per_response"`
	XpReward          int          `json:"xp_reward"`
	EstimatedDuration int          `json:"estimated_duration"`
	ResponseCount     int          `json:"response_count"`
	MaxResponses      int          `json:"max_responses"`
	CompletionRate    float64      `json:"completion_rate"`
	AverageRating     float64      `json:"average_rating"`
	CreatedAt         time.Time    `json:"created_at"`
	Creator           UserResponse `json:"creator"`
	Progress          float64      `json:"progress"`
}