package schemas

import "time"

// Signer is a participant asked to act on a document, holding its role, status
// and the token that grants access to the signing page. Token is sealed at
// rest when PORTE_TOKEN_KEY is set; TokenHash is the SHA-256 hex every lookup
// goes through.
type Signer struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	DocumentID     int64      `gorm:"column:document_id;index"`
	Name           string     `gorm:"column:name"`
	Email          string     `gorm:"column:email"`
	Role           string     `gorm:"column:role;default:signer"`
	Status         string     `gorm:"column:status;default:pending"`
	Token          string     `gorm:"column:token;index"`
	TokenHash      string     `gorm:"column:token_hash;index"`
	OrderNum       int        `gorm:"column:order_num;default:0"`
	SignedAt       *time.Time `gorm:"column:signed_at"`
	ViewedAt       *time.Time `gorm:"column:viewed_at"`
	EmailOpenedAt  *time.Time `gorm:"column:email_opened_at"`
	IPAddress      string     `gorm:"column:ip_address"`
	UserAgent      string     `gorm:"column:user_agent"`
	LastRemindedAt *time.Time `gorm:"column:last_reminded_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (Signer) TableName() string { return "signers" }
