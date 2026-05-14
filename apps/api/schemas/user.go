package schemas

import "time"

type User struct {
	ID                   int64     `gorm:"column:id;primaryKey"`
	Email                string    `gorm:"column:email;uniqueIndex"`
	Name                 string    `gorm:"column:name"`
	PasswordHash         string    `gorm:"column:password_hash"`
	ReminderIntervalDays int       `gorm:"column:reminder_interval_days;default:3"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (User) TableName() string { return "users" }
