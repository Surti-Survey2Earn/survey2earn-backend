package models

import (
	"time"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type SoftDelete interface {
	IsDeleted() bool
}

type Timestamped interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

func (b BaseModel) GetCreatedAt() time.Time {
	return b.CreatedAt
}

func (b BaseModel) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

func (b BaseModel) IsDeleted() bool {
	return b.DeletedAt.Valid
}