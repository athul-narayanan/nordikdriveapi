package role

type Role struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Role          string `gorm:"unique;not null" json:"role"`
	Priority      uint   `gorm:"not null" json:"priority"`
	CanUpload     bool   `gorm:"not null" json:"can_upload"`
	CanView       bool   `gorm:"not null" json:"can_view"`
	CanApprove    bool   `gorm:"not null" json:"can_approve"`
	CanApproveAll bool   `gorm:"not null" json:"can_approve_all"`
}
