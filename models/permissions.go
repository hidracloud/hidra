package models

import (
	"fmt"

	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Permission represents a permission
type Permission struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primaryKey;type:char(36);"`
	UserID  uuid.UUID `json:"user"`
	User    User      `gorm:"foreignKey:UserID" json:"-"`
	AllowTo string
}

// AddPermission2User add permission to user
func AddPermission2User(user *User, allowTo string) (*Permission, error) {
	newPermission := Permission{ID: uuid.NewV4(), AllowTo: allowTo, User: *user}
	if result := database.ORM.Create(&newPermission); result.Error != nil {
		return nil, result.Error
	}
	return &newPermission, nil
}

// GetPermissionByUserAllowTo get permission by user and allowTo
func GetPermissionByUserAllowTo(user *User, allowTo string) (*Permission, error) {
	var permission Permission
	database.ORM.First(&permission, "user_id = ? AND allow_to = ?", user.ID, allowTo)

	if permission.ID == uuid.Nil {
		return nil, fmt.Errorf("user has not permission to execute")
	}

	return &permission, nil
}

// CheckIfAllowTo check if user has permission to execute
func CheckIfAllowTo(user *User, allowTo string) error {
	_, err := GetPermissionByUserAllowTo(user, "superadmin")

	if err == nil {
		return nil
	}

	_, err = GetPermissionByUserAllowTo(user, allowTo)
	return err
}
