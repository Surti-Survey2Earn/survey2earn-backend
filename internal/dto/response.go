// internal/dto/response.go
package dto

import (
	"time"
)

// StartSurveyRequest represents the request to start a survey
type StartSurveyRequest struct {
	SurveyID  uint   `json:"survey_id" binding:"required"`
	Timezone  string `json:"timezone"`
	Language  string `json:"language"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// SubmitAnswerRequest represents a single answer submission
type SubmitAnswerRequest struct {
	QuestionID uint        `json:"question_id" binding:"required"`
	Answer     AnswerValue `json:"answer" binding:"required"`
	TimeSpent  int         `json:"time_spent"` // in seconds
	IsSkipped  bool        `json:"is_skipped"`
}

// AnswerValue represents the answer value structure
type AnswerValue struct {
	Type     string      `json:"type" binding:"required"`    // text, number, array, boolean, rating, scale, date
	Content  interface{} `json:"value"`                      // The actual answer value
	Options  []string    `json:"options"`                    // Selected options for multiple choice
	Rating   *int        `json:"rating"`                     // Rating value (1-5)
	Scale    *int        `json:"scale"`                      // Scale value (1-10)
	Date     *time.Time  `json:"date"`                       // Date value
}

// CompleteSurveyRequest represents the final survey submission
type CompleteSurveyRequest struct {
	ResponseID uint                      `json:"response_id" binding:"required"`
	Answers    []SubmitAnswerRequest     `json:"answers" binding:"required"`
	Duration   int                       `json:"duration"` // total time spent in seconds
}

// UpdateAnswerRequest for updating a single answer
type UpdateAnswerRequest struct {
	Answer    AnswerValue `json:"answer" binding:"required"`
	TimeSpent int         `json:"time_spent"`
	IsSkipped bool        `json:"is_skipped"`
}

// ResponseStartResponse represents the response when starting a survey
type ResponseStartResponse struct {
	ResponseID uint      `json:"response_id"`
	SurveyID   uint      `json:"survey_id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	TimeLeft   *int      `json:"time_left"` // in seconds, if survey has time limit
}

// AnswerResponse represents an answer in response
type AnswerResponse struct {
	ID         uint        `json:"id"`
	QuestionID uint        `json:"question_id"`
	Answer     AnswerValue `json:"answer"`
	TimeSpent  int         `json:"time_spent"`
	IsSkipped  bool        `json:"is_skipped"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// SurveyResponseResponse represents the complete survey response
type SurveyResponseResponse struct {
	ID            uint             `json:"id"`
	SurveyID      uint             `json:"survey_id"`
	UserID        uint             `json:"user_id"`
	Status        string           `json:"status"`
	StartedAt     time.Time        `json:"started_at"`
	CompletedAt   *time.Time       `json:"completed_at"`
	Duration      int              `json:"duration"`
	QualityScore  float64          `json:"quality_score"`
	IsValid       bool             `json:"is_valid"`
	Answers       []AnswerResponse `json:"answers"`
	RewardEarned  float64          `json:"reward_earned"`
	XpEarned      int              `json:"xp_earned"`
	NFTCertificate *string         `json:"nft_certificate"`
}

// CompletionResponse represents the response after completing a survey
type CompletionResponse struct {
	ResponseID      uint      `json:"response_id"`
	Status          string    `json:"status"`
	CompletedAt     time.Time `json:"completed_at"`
	Duration        int       `json:"duration"`
	RewardEarned    float64   `json:"reward_earned"`
	XpEarned        int       `json:"xp_earned"`
	NFTCertificate  *string   `json:"nft_certificate"`
	TransactionHash *string   `json:"transaction_hash"`
	Message         string    `json:"message"`
}

// SurveyProgressResponse for tracking survey progress
type SurveyProgressResponse struct {
	ResponseID        uint      `json:"response_id"`
	SurveyID          uint      `json:"survey_id"`
	Status            string    `json:"status"`
	Progress          float64   `json:"progress"` // percentage (0-100)
	QuestionsTotal    int       `json:"questions_total"`
	QuestionsAnswered int       `json:"questions_answered"`
	TimeSpent         int       `json:"time_spent"`
	TimeLeft          *int      `json:"time_left"`
	StartedAt         time.Time `json:"started_at"`
	LastAnsweredAt    *time.Time `json:"last_answered_at"`
}

// ListResponsesRequest for filtering user responses
type ListResponsesRequest struct {
	Status    string `form:"status"`
	SurveyID  uint   `form:"survey_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page" binding:"min=1"`
	Limit     int    `form:"limit" binding:"min=1,max=100"`
}

// ResponseListResponse for listing user responses
type ResponseListResponse struct {
	Responses  []ResponseItemResponse `json:"responses"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

// ResponseItemResponse for response list item
type ResponseItemResponse struct {
	ID            uint         `json:"id"`
	SurveyID      uint         `json:"survey_id"`
	SurveyTitle   string       `json:"survey_title"`
	Status        string       `json:"status"`
	StartedAt     time.Time    `json:"started_at"`
	CompletedAt   *time.Time   `json:"completed_at"`
	Duration      int          `json:"duration"`
	RewardEarned  float64      `json:"reward_earned"`
	XpEarned      int          `json:"xp_earned"`
	QualityScore  float64      `json:"quality_score"`
	Progress      float64      `json:"progress"`
}