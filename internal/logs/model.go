package logs

import (
	"time"
)

type SystemLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Level     string    `gorm:"size:20;not null" json:"level"`
	Service   string    `gorm:"size:100;not null" json:"service"`
	UserID    *uint     `gorm:"index" json:"user_id,omitempty"`
	Action    string    `gorm:"size:255;not null" json:"action"`
	Message   string    `gorm:"type:text" json:"message"`
	Metadata  *string   `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type LogFilterInput struct {
	UserID    *uint   `json:"user_id"`
	Level     *string `json:"level"`
	Service   *string `json:"service"`
	StartDate *string `json:"start_date"` // "YYYY-MM-DD"
	EndDate   *string `json:"end_date"`
	Search    *string `json:"search"`
	Page      int     `json:"page"`
	Action    *string `json:"action"`
	PageSize  int     `json:"page_size"`
}

func (SystemLog) TableName() string {
	return "logs"
}
