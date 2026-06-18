package schemas

import "time"

type Space struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	OwnerID     int64     `gorm:"column:owner_id;index"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Space) TableName() string { return "spaces" }

type SpaceMember struct {
	ID       int64     `gorm:"column:id;primaryKey"`
	SpaceID  int64     `gorm:"column:space_id;index"`
	UserID   int64     `gorm:"column:user_id;index"`
	Role     string    `gorm:"column:role;default:member"`
	JoinedAt time.Time `gorm:"column:joined_at;autoCreateTime"`

	Space Space `gorm:"foreignKey:SpaceID;constraint:OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (SpaceMember) TableName() string { return "space_members" }
