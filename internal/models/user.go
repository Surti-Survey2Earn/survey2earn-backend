package models

import (
	"strings"
	"time"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	WalletAddress  string    `json:"wallet_address" gorm:"unique;not null;index"`
	Nonce          string    `json:"-" gorm:"not null"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	
	Username       *string   `json:"username" gorm:"unique"`
	Email          *string   `json:"email" gorm:"unique"`
	ProfilePicture *string   `json:"profile_picture"`
	Bio            *string   `json:"bio" gorm:"type:text"`
	
	ReputationScore float64  `json:"reputation_score" gorm:"default:0"`
	TotalEarned     float64  `json:"total_earned" gorm:"default:0"`
	TotalResponses  int      `json:"total_responses" gorm:"default:0"`
	TotalSurveys    int      `json:"total_surveys" gorm:"default:0"`
	
	Surveys         []Survey         `json:"surveys,omitempty" gorm:"foreignKey:CreatorID"`
	Responses       []Response       `json:"responses,omitempty" gorm:"foreignKey:UserID"`
	AuthSessions    []AuthSession    `json:"-" gorm:"foreignKey:UserID"`
	Transactions    []RewardTransaction `json:"transactions,omitempty" gorm:"foreignKey:UserID"`
}

type AuthSession struct {
	BaseModel
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"unique;not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

type UserStats struct {
	UserID              uint    `json:"user_id" gorm:"primaryKey"`
	TotalSurveysCreated int     `json:"total_surveys_created" gorm:"default:0"`
	TotalSurveysAnswered int    `json:"total_surveys_answered" gorm:"default:0"`
	TotalEarned         float64 `json:"total_earned" gorm:"default:0"`
	TotalSpent          float64 `json:"total_spent" gorm:"default:0"`
	AverageRating       float64 `json:"average_rating" gorm:"default:0"`
	LastActivityAt      *time.Time `json:"last_activity_at"`
	
	User                User    `json:"user" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.WalletAddress = strings.ToLower(u.WalletAddress)
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.WalletAddress = strings.ToLower(u.WalletAddress)
	return nil
}

func (as *AuthSession) IsSessionValid() bool {
	return as.IsActive && time.Now().Before(as.ExpiresAt)
}

func (User) TableName() string {
	return "users"
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}

func (UserStats) TableName() string {
	return "user_stats"
}