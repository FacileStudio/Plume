package schemas

import "time"

// Document is a signature-bearing PDF and its workflow state, from draft
// through pending to completed or declined.
//
// The last four columns serve apps that ask Plume for signatures: the caller's
// own reference, the Nuage space and idempotency key to archive the signed PDF
// under, and when that archive succeeded. All are empty for a document created
// through the UI.
type Document struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	Name         string    `gorm:"column:name"`
	Status       string    `gorm:"column:status;default:draft;index"`
	FileName     string    `gorm:"column:file_name"`
	StoragePath  string    `gorm:"column:storage_path"`
	OriginalHash string    `gorm:"column:original_hash;index"`
	SignedHash   string    `gorm:"column:signed_hash;index"`
	OwnerID      int64     `gorm:"column:owner_id;index"`
	SpaceID      *int64    `gorm:"column:space_id;index"`
	ClientID     *int64    `gorm:"column:client_id;index"`
	Sequential   bool      `gorm:"column:sequential;default:false"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`

	ExternalRef        string     `gorm:"column:external_ref;size:200;not null;default:''"`
	NuageSpaceFacileID string     `gorm:"column:nuage_space_facile_id;size:40;not null;default:''"`
	NuageLinkedTo      string     `gorm:"column:nuage_linked_to;size:200;not null;default:''"`
	NuagePublishedAt   *time.Time `gorm:"column:nuage_published_at"`
}

func (Document) TableName() string { return "documents" }
