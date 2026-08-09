package schemas

import "time"

// Field carries no `default` tag on Required on purpose: GORM omits a zero-valued column
// from the INSERT when the tag declares a default, so `default:true` silently stored every
// optional field as required, and the completion guard would then reject a signer for
// leaving a box the sender marked optional.
type Field struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	DocumentID int64     `gorm:"column:document_id;index"`
	SignerID   int64     `gorm:"column:signer_id;index"`
	FieldType  string    `gorm:"column:field_type"`
	Page       int       `gorm:"column:page"`
	X          float64   `gorm:"column:x"`
	Y          float64   `gorm:"column:y"`
	Width      float64   `gorm:"column:width"`
	Height     float64   `gorm:"column:height"`
	Required   bool      `gorm:"column:required"`
	Label      string    `gorm:"column:label"`
	Value      string    `gorm:"column:value"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Field) TableName() string { return "fields" }
