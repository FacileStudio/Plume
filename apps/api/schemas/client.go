package schemas

import "time"

type Client struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	Email     string    `gorm:"column:email"`
	Company   string    `gorm:"column:company"`
	Phone     string    `gorm:"column:phone"`
	Notes     string    `gorm:"column:notes"`
	OwnerID   int64     `gorm:"column:owner_id;index"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Client) TableName() string { return "clients" }
