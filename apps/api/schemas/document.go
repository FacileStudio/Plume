package schemas

import "time"

type Document struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Status      string    `gorm:"column:status;default:draft;index"`
	FileName    string    `gorm:"column:file_name"`
	StoragePath string    `gorm:"column:storage_path"`
	OwnerID     int64     `gorm:"column:owner_id;index"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Document) TableName() string { return "documents" }
