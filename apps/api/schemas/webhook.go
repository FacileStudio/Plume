package schemas

import "time"

// Webhook is an outbound delivery target for document and signer lifecycle
// events, authenticated with an HMAC secret.
type Webhook struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	OwnerID    int64      `gorm:"column:owner_id;index"`
	SpaceID    *int64     `gorm:"column:space_id;index"`
	URL        string     `gorm:"column:url"`
	Secret     string     `gorm:"column:secret"`
	Enabled    bool       `gorm:"column:enabled;default:true"`
	LastSentAt *time.Time `gorm:"column:last_sent_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Webhook) TableName() string { return "webhooks" }
