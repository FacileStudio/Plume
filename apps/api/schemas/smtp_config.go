package schemas

import "time"

type SmtpConfig struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	OwnerID   int64     `gorm:"column:owner_id;uniqueIndex"`
	Host      string    `gorm:"column:host"`
	Port      int       `gorm:"column:port"`
	Username  string    `gorm:"column:username"`
	Password  string    `gorm:"column:password"`
	FromEmail string    `gorm:"column:from_email"`
	FromName  string    `gorm:"column:from_name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SmtpConfig) TableName() string { return "smtp_configs" }
