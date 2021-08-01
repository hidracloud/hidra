package models

import (
	"fmt"

	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Represent user permissions
type Permission struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primaryKey;type:char(36);"`
	UserId  uuid.UUID `json:"user"`
	User    User      `gorm:"foreignKey:UserId" json:"-"`
	AllowTo string
}

// Add new permission to user
func AddPermission2User(user *User, allowTo string) (*Permission, error) {
	newPermission := Permission{ID: uuid.NewV4(), AllowTo: allowTo, User: *user}
	if result := database.ORM.Create(&newPermission); result.Error != nil {
		return nil, result.Error
	}
	return &newPermission, nil
}

// Get permission permission by allow to and by user
func GetPermissionByUserAllowTo(user *User, allowTo string) (*Permission, error) {
	var permission Permission
	database.ORM.First(&permission, "user_id = ? AND allow_to = ?", user.ID, allowTo)

	if permission.ID == uuid.Nil {
		return nil, fmt.Errorf("user has not permission to execute")
	}

	return &permission, nil
}

// Check if current user is allow to do an action
func CheckIfAllowTo(user *User, allowTo string) error {
	_, err := GetPermissionByUserAllowTo(user, "superadmin")

	if err == nil {
		return nil
	}

	_, err = GetPermissionByUserAllowTo(user, allowTo)
	return err
}
