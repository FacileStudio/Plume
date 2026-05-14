package schemas

import "time"

type Signer struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	DocumentID int64      `gorm:"column:document_id;index"`
	Name       string     `gorm:"column:name"`
	Email      string     `gorm:"column:email"`
	Role       string     `gorm:"column:role;default:signer"`
	Status     string     `gorm:"column:status;default:pending"`
	Token      string     `gorm:"column:token;index"`
	OrderNum   int        `gorm:"column:order_num;default:0"`
	SignedAt       *time.Time `gorm:"column:signed_at"`
	IPAddress      string     `gorm:"column:ip_address"`
	UserAgent      string     `gorm:"column:user_agent"`
	LastRemindedAt *time.Time `gorm:"column:last_reminded_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (Signer) TableName() string { return "signers" }
