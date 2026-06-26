package model

import "time"

type Role struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name;size:64;not null"`
	Code      string    `gorm:"column:code;size:64;uniqueIndex;not null"`
	Remark    string    `gorm:"column:remark;size:255"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Role) TableName() string {
	return "roles"
}