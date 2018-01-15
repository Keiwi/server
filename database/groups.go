package database

import (
	"time"

	// import mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Group struct
type Group struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
	CommandID uint       `gorm:"command_id" json:"command_id"`
	GroupName string     `gorm:"not null" json:"group_name"`
	NextCheck int        `json:"next_check"`
	StopError bool       `json:"stop_error"`
}

// GroupsManager struct
type GroupsManager struct {
	db *DB
}

// NewGroupsManager - Creates a new *GroupsManager that can be used for managing groups.
func NewGroupsManager(db *DB) (*GroupsManager, error) {
	db.AutoMigrate(&Group{})
	manager := GroupsManager{}
	manager.db = db
	return &manager, nil
}

// FindAll - returns all groups in database
func (state GroupsManager) FindAll() ([]*Group, error) {
	groups := []*Group{}
	err := state.db.Find(&groups).Error
	return groups, err
}

// Find - returns a specific client based on id in datbase
func (state GroupsManager) Find(id uint) (*Group, error) {
	group := &Group{}
	err := state.db.Where("id = ?", id).Find(group).Error
	return group, err
}

// FindWithName - returns a specific client based on name in datbase
func (state GroupsManager) FindWithName(name string) ([]*Group, error) {
	groups := []*Group{}
	err := state.db.Where("group_name = ?", name).Find(groups).Error
	return groups, err
}

// FindWithCommand - returns a specific client based on command in datbase
func (state GroupsManager) FindWithCommand(name string, id uint) (*Group, error) {
	group := &Group{}
	err := state.db.Where("group_name = ? AND command_id = ?", name, id).Find(group).Error
	return group, err
}
