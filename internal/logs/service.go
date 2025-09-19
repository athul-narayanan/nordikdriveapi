package logs

import (
	"encoding/json"
	"math"
	"time"

	"gorm.io/gorm"
)

type LogService struct {
	DB *gorm.DB
}

func (ls *LogService) Log(level, service, action, message string, userID *uint, metadata interface{}) error {
	var metaStr *string

	// Convert metadata (map/struct) to JSON string if provided
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			str := string(b)
			metaStr = &str
		}
	}

	log := SystemLog{
		Level:     level,
		Service:   service,
		UserID:    userID,
		Action:    action,
		Message:   message,
		Metadata:  metaStr,
		CreatedAt: time.Now(),
	}

	return ls.DB.Create(&log).Error

}

func (ls *LogService) GetLogs(input LogFilterInput) ([]map[string]interface{}, int64, int, error) {
	// Defaults
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 || input.PageSize > 100 {
		input.PageSize = 20
	}

	db := ls.DB.Model(&SystemLog{}).
		Select("logs.*, a.firstname, a.lastname").
		Joins("LEFT JOIN users a ON logs.user_id = a.id")

	// Default: last 30 days
	if input.StartDate == nil && input.EndDate == nil {
		db = db.Where("logs.created_at >= ?", time.Now().AddDate(0, 0, -30))
	}

	// Apply filters
	if input.UserID != nil {
		db = db.Where("logs.user_id = ?", *input.UserID)
	}
	if input.Level != nil {
		db = db.Where("logs.level = ?", *input.Level)
	}

	if input.Service != nil {
		db = db.Where("logs.service = ?", *input.Service)
	}

	if input.Action != nil {
		db = db.Where("logs.action = ?", *input.Action)
	}

	// Date range
	if input.StartDate != nil && input.EndDate != nil {
		db = db.Where("logs.created_at BETWEEN ? AND ?", *input.StartDate, *input.EndDate)
	} else if input.StartDate != nil {
		db = db.Where("logs.created_at >= ?", *input.StartDate)
	} else if input.EndDate != nil {
		db = db.Where("logs.created_at <= ?", *input.EndDate)
	}

	// Search across multiple columns
	if input.Search != nil && *input.Search != "" {
		like := "%" + *input.Search + "%"
		db = db.Where(
			"CAST(logs.id AS TEXT) ILIKE ? OR logs.level ILIKE ? OR logs.service ILIKE ? OR logs.action ILIKE ? OR logs.message ILIKE ? OR a.firstname ILIKE ? OR a.lastname ILIKE ?",
			like, like, like, like, like, like, like,
		)
	}

	// Count total
	var total int64
	db.Count(&total)

	// Pagination + query
	var logs []map[string]interface{}
	if err := db.
		Limit(input.PageSize).
		Offset((input.Page - 1) * input.PageSize).
		Order("logs.created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(input.PageSize)))
	return logs, total, totalPages, nil
}
