package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// SurveyStatus represents the status of a survey
type SurveyStatus string

const (
	SurveyStatusDraft     SurveyStatus = "draft"
	SurveyStatusPublished SurveyStatus = "published"
	SurveyStatusPaused    SurveyStatus = "paused"
	SurveyStatusCompleted SurveyStatus = "completed"
	SurveyStatusCancelled SurveyStatus = "cancelled"
)

// QuestionType represents the type of question
type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeText           QuestionType = "text"
	QuestionTypeTextArea       QuestionType = "textarea"
	QuestionTypeRating         QuestionType = "rating"
	QuestionTypeYesNo          QuestionType = "yes_no"
	QuestionTypeScale          QuestionType = "scale"
	QuestionTypeDate           QuestionType = "date"
	QuestionTypeNumber         QuestionType = "number"
)

// Survey represents a survey
type Survey struct {
	BaseModel
	CreatorID         uint           `json:"creator_id" gorm:"not null;index"`
	Title             string         `json:"title" gorm:"not null;size:255"`
	Description       string         `json:"description" gorm:"type:text"`
	Category          string         `json:"category" gorm:"not null;size:100;index"`
	Status            SurveyStatus   `json:"status" gorm:"default:'draft';index"`
	
	// Survey Configuration
	MaxResponses      int            `json:"max_responses" gorm:"default:100"`
	MinResponses      int            `json:"min_responses" gorm:"default:1"`
	RewardPerResponse float64        `json:"reward_per_response" gorm:"not null"`
	TotalRewardPool   float64        `json:"total_reward_pool" gorm:"not null"`
	
	// Time Configuration
	StartDate         *time.Time     `json:"start_date"`
	EndDate           *time.Time     `json:"end_date"`
	EstimatedDuration int            `json:"estimated_duration"` // in minutes
	
	// Survey Settings
	IsAnonymous       bool           `json:"is_anonymous" gorm:"default:true"`
	IsPublic          bool           `json:"is_public" gorm:"default:true"`
	RequireLogin      bool           `json:"require_login" gorm:"default:true"`
	AllowMultiple     bool           `json:"allow_multiple" gorm:"default:false"`
	
	// Statistics
	ResponseCount     int            `json:"response_count" gorm:"default:0"`
	CompletionRate    float64        `json:"completion_rate" gorm:"default:0"`
	AverageRating     float64        `json:"average_rating" gorm:"default:0"`
	
	// Relationships
	Creator           User           `json:"creator" gorm:"foreignKey:CreatorID"`
	Questions         []Question     `json:"questions" gorm:"foreignKey:SurveyID;constraint:OnDelete:CASCADE"`
	Responses         []Response     `json:"responses,omitempty" gorm:"foreignKey:SurveyID"`
	RewardPool        *RewardPool    `json:"reward_pool,omitempty" gorm:"foreignKey:SurveyID"`
}

// Question represents a question in a survey
type Question struct {
	BaseModel
	SurveyID     uint               `json:"survey_id" gorm:"not null;index"`
	Type         QuestionType       `json:"type" gorm:"not null"`
	Text         string             `json:"text" gorm:"not null;type:text"`
	Description  string             `json:"description" gorm:"type:text"`
	Options      QuestionOptions    `json:"options" gorm:"type:json"`
	Required     bool               `json:"required" gorm:"default:false"`
	Order        int                `json:"order" gorm:"not null"`
	
	// Question Configuration
	MinLength    *int               `json:"min_length"`
	MaxLength    *int               `json:"max_length"`
	MinValue     *float64           `json:"min_value"`
	MaxValue     *float64           `json:"max_value"`
	
	// Conditional Logic
	ShowIf       *ConditionalLogic  `json:"show_if" gorm:"type:json"`
	
	// Relationships
	Survey       Survey             `json:"survey" gorm:"foreignKey:SurveyID"`
	Answers      []Answer           `json:"answers,omitempty" gorm:"foreignKey:QuestionID"`
}

// QuestionOptions represents the options for a question
type QuestionOptions []QuestionOption

type QuestionOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
	Order int    `json:"order"`
}

// ConditionalLogic represents conditional logic for showing questions
type ConditionalLogic struct {
	QuestionID string      `json:"question_id"`
	Operator   string      `json:"operator"` // equals, not_equals, contains, greater_than, less_than
	Value      interface{} `json:"value"`
}

// QuestionOptionsValue implements driver.Valuer interface for QuestionOptions
func (qo QuestionOptions) Value() (driver.Value, error) {
	return json.Marshal(qo)
}

// Scan implements sql.Scanner interface for QuestionOptions
func (qo *QuestionOptions) Scan(value interface{}) error {
	if value == nil {
		*qo = QuestionOptions{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-bytes into QuestionOptions")
	}
	
	return json.Unmarshal(bytes, qo)
}

// ConditionalLogicValue implements driver.Valuer interface for ConditionalLogic
// func (cl ConditionalLogic) Value() (driver.Value, error) {
// 	return json.Marshal(cl)
// }

// Scan implements sql.Scanner interface for ConditionalLogic
func (cl *ConditionalLogic) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-bytes into ConditionalLogic")
	}
	
	return json.Unmarshal(bytes, cl)
}

// IsActive checks if the survey is currently active
func (s *Survey) IsActive() bool {
	now := time.Now()
	
	if s.Status != SurveyStatusPublished {
		return false
	}
	
	if s.StartDate != nil && now.Before(*s.StartDate) {
		return false
	}
	
	if s.EndDate != nil && now.After(*s.EndDate) {
		return false
	}
	
	if s.ResponseCount >= s.MaxResponses {
		return false
	}
	
	return true
}

// CanBeEdited checks if the survey can be edited
func (s *Survey) CanBeEdited() bool {
	return s.Status == SurveyStatusDraft
}

// CalculateTotalRewardPool calculates the total reward pool needed
func (s *Survey) CalculateTotalRewardPool() float64 {
	return float64(s.MaxResponses) * s.RewardPerResponse
}

// GetQuestionByID finds a question by its ID
func (s *Survey) GetQuestionByID(questionID uint) (*Question, error) {
	for _, question := range s.Questions {
		if question.ID == questionID {
			return &question, nil
		}
	}
	return nil, errors.New("question not found")
}

// TableName returns the table name for Survey
func (Survey) TableName() string {
	return "surveys"
}

// TableName returns the table name for Question
func (Question) TableName() string {
	return "questions"
}