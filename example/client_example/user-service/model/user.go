package model

import "time"

// User 对应 PostgreSQL 中的 users 表，gorm 自动建表
type User struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	Username  string    `gorm:"column:username;size:64;uniqueIndex"`
	Password  string    `gorm:"column:password;size:128"`
	RoleID    uint64    `gorm:"column:role_id"`
	Email     string    `gorm:"column:email;size:128"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
