package file

import (
	"time"

	"gorm.io/datatypes"
)

type File struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Filename   string    `gorm:"unique;not null" json:"filename"`
	InsertedBy uint      `gorm:"not null" json:"inserted_by"`
	CreatedAt  time.Time `json:"created_at"`
	Private    bool      `json:"private"`
	IsDelete   bool      `json:"is_delete"`
	Size       float64   `json:"size"`
	Version    int       `json:"version"`
	Rows       int       `json:"rows"`
}

type FileVersion struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" jsoxn:"id"`
	FileID     uint      `gorm:"not null;index" json:"file_id"`
	Filename   string    `gorm:"size:255;unique;not null" json:"filename"`
	InsertedBy uint      `gorm:"not null;index" json:"inserted_by"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	Private    bool      `gorm:"default:false" json:"private"`
	IsDelete   bool      `gorm:"default:false" json:"is_delete"`
	Size       float64   `gorm:"not null" json:"size"`
	Version    int       `gorm:"not null;default:1" json:"version"`
	Rows       int       `gorm:"not null" json:"rows"`
}

type FileVersionWithUser struct {
	ID        uint      `json:"id"`
	FileID    uint      `json:"file_id"`
	FileName  string    `json:"filename" gorm:"column:filename"`
	Firstname string    `json:"firstname" gorm:"column:firstname"`
	Lastname  string    `json:"lastname" gorm:"column:lastname"`
	CreatedAt time.Time `json:"created_at"`
	Private   bool      `json:"private"`
	IsDelete  bool      `json:"is_delete"`
	Size      float64   `json:"size"`
	Version   int       `json:"version"`
	Rows      int       `json:"rows"`
}

type FileData struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileID     uint           `gorm:"not null;index" json:"file_id"`
	RowData    datatypes.JSON `gorm:"type:jsonb" json:"row_data"`
	InsertedBy uint           `gorm:"not null" json:"inserted_by"`
	CreatedAt  time.Time      `json:"created_at"`
	Version    int            `json:"version"`
}

type FileAccess struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"type:json" json:"user_id"`
	FileID uint `gorm:"not null;index" json:"file_id"`
}

type RevertFileInput struct {
	Filename string `json:"filename" binding:"required"`
	Version  int    `json:"version" binding:"required"`
}

type FileWithUser struct {
	File
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

func (File) TableName() string {
	return "file"
}

func (FileData) TableName() string {
	return "file_data"
}

func (FileAccess) TableName() string {
	return "file_access"
}

func (FileVersion) TableName() string {
	return "file_version"
}
