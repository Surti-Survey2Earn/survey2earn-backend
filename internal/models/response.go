package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ResponseStatus represents the status of a response
type ResponseStatus string

const (
	ResponseStatusStarted   ResponseStatus = "started"
	ResponseStatusCompleted ResponseStatus = "completed"
	ResponseStatusAbandoned ResponseStatus = "abandoned"
)

// Response represents a user's response to a survey
type Response struct {
	BaseModel
	SurveyID      uint             `json:"survey_id" gorm:"not null;index"`
	UserID        uint             `json:"user_id" gorm:"not null;index"`
	Status        ResponseStatus   `json:"status" gorm:"default:'started';index"`
	
	// Timing Information
	StartedAt     time.Time        `json:"started_at" gorm:"not null"`
	CompletedAt   *time.Time       `json:"completed_at"`
	Duration      int              `json:"duration"` // in seconds
	
	// Response Metadata
	IPAddress     string           `json:"ip_address"`
	UserAgent     string           `json:"user_agent"`
	Timezone      string           `json:"timezone"`
	Language      string           `json:"language" gorm:"default:'en'"`
	
	// Quality Metrics
	QualityScore  float64          `json:"quality_score" gorm:"default:0"`
	IsValid       bool             `json:"is_valid" gorm:"default:true"`
	FlaggedReason *string          `json:"flagged_reason"`
	
	// Relationships
	Survey        Survey           `json:"survey" gorm:"foreignKey:SurveyID"`
	User          User             `json:"user" gorm:"foreignKey:UserID"`
	Answers       []Answer         `json:"answers" gorm:"foreignKey:ResponseID;constraint:OnDelete:CASCADE"`
	Transaction   *RewardTransaction `json:"transaction,omitempty" gorm:"foreignKey:ResponseID"`
}

// Answer represents an answer to a specific question
type Answer struct {
	BaseModel
	ResponseID    uint             `json:"response_id" gorm:"not null;index"`
	QuestionID    uint             `json:"question_id" gorm:"not null;index"`
	
	// Answer Data
	AnswerText    string           `json:"answer_text" gorm:"type:text"`
	AnswerValue   AnswerValue      `json:"answer_value" gorm:"type:json"`
	
	// Answer Metadata
	TimeSpent     int              `json:"time_spent"` // in seconds
	IsSkipped     bool             `json:"is_skipped" gorm:"default:false"`
	
	// Relationships
	Response      Response         `json:"response" gorm:"foreignKey:ResponseID"`
	Question      Question         `json:"question" gorm:"foreignKey:QuestionID"`
}

// AnswerValue represents the structured value of an answer
type AnswerValue struct {
	Type       string      `json:"type"`     // text, number, array, boolean
	Content    interface{} `json:"value"`    // The actual answer value (keep JSON tag as "value")
	Options    []string    `json:"options"`  // Selected options for multiple choice
	Rating     *int        `json:"rating"`   // Rating value
	Scale      *int        `json:"scale"`    // Scale value
	Date       *time.Time  `json:"date"`     // Date value
}

// ResponseSummary represents a summary of responses for analytics
type ResponseSummary struct {
	SurveyID         uint      `json:"survey_id" gorm:"primaryKey"`
	TotalResponses   int       `json:"total_responses" gorm:"default:0"`
	CompletedCount   int       `json:"completed_count" gorm:"default:0"`
	AbandonedCount   int       `json:"abandoned_count" gorm:"default:0"`
	AverageDuration  float64   `json:"average_duration" gorm:"default:0"`
	CompletionRate   float64   `json:"completion_rate" gorm:"default:0"`
	AverageQuality   float64   `json:"average_quality" gorm:"default:0"`
	LastResponseAt   *time.Time `json:"last_response_at"`
	
	// Relationship
	Survey           Survey    `json:"survey" gorm:"foreignKey:SurveyID"`
}

// Value implements driver.Valuer interface for AnswerValue
func (av AnswerValue) Value() (driver.Value, error) {
	return json.Marshal(av)
}

// Scan implements sql.Scanner interface for AnswerValue
func (av *AnswerValue) Scan(value interface{}) error {
	if value == nil {
		*av = AnswerValue{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-bytes into AnswerValue")
	}
	
	return json.Unmarshal(bytes, av)
}

// IsCompleted checks if the response is completed
func (r *Response) IsCompleted() bool {
	return r.Status == ResponseStatusCompleted && r.CompletedAt != nil
}

// CalculateDuration calculates the duration of the response
func (r *Response) CalculateDuration() int {
	if r.CompletedAt != nil {
		return int(r.CompletedAt.Sub(r.StartedAt).Seconds())
	}
	return int(time.Since(r.StartedAt).Seconds())
}

// MarkAsCompleted marks the response as completed
func (r *Response) MarkAsCompleted() {
	now := time.Now()
	r.Status = ResponseStatusCompleted
	r.CompletedAt = &now
	r.Duration = r.CalculateDuration()
}

// MarkAsAbandoned marks the response as abandoned
func (r *Response) MarkAsAbandoned() {
	r.Status = ResponseStatusAbandoned
	r.Duration = r.CalculateDuration()
}

// GetAnswerByQuestionID finds an answer by question ID
func (r *Response) GetAnswerByQuestionID(questionID uint) (*Answer, error) {
	for _, answer := range r.Answers {
		if answer.QuestionID == questionID {
			return &answer, nil
		}
	}
	return nil, errors.New("answer not found")
}

// ValidateAnswer validates an answer based on question requirements
func (a *Answer) ValidateAnswer(question *Question) error {
	if question.Required && (a.IsSkipped || a.AnswerText == "") {
		return errors.New("answer is required")
	}
	
	// Additional validation based on question type
	switch question.Type {
	case QuestionTypeText, QuestionTypeTextArea:
		if question.MinLength != nil && len(a.AnswerText) < *question.MinLength {
			return errors.New("answer too short")
		}
		if question.MaxLength != nil && len(a.AnswerText) > *question.MaxLength {
			return errors.New("answer too long")
		}
	case QuestionTypeRating, QuestionTypeScale:
		if a.AnswerValue.Rating != nil {
			if question.MinValue != nil && float64(*a.AnswerValue.Rating) < *question.MinValue {
				return errors.New("rating below minimum")
			}
			if question.MaxValue != nil && float64(*a.AnswerValue.Rating) > *question.MaxValue {
				return errors.New("rating above maximum")
			}
		}
	}
	
	return nil
}

// TableName returns the table name for Response
func (Response) TableName() string {
	return "responses"
}

// TableName returns the table name for Answer
func (Answer) TableName() string {
	return "answers"
}

// TableName returns the table name for ResponseSummary
func (ResponseSummary) TableName() string {
	return "response_summaries"
}