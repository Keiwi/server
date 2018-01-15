package database

import (
	"time"

	// import mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Check struct
type Check struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
	CommandID uint       `json:"command_id"`
	ClientID  uint       `json:"client_id"`
	Response  string     `gorm:"not null" json:"response"`
	Checked   bool       `json:"checked"`
	Error     bool       `json:"error"`
	Finished  bool       `json:"finished"`
}

// ChecksManager struct
type ChecksManager struct {
	db *DB
}

// NewChecksManager - Creates a new *ChecksManager that can be used for managing checks.
func NewChecksManager(db *DB) (*ChecksManager, error) {
	db.AutoMigrate(&Check{})
	manager := ChecksManager{}
	manager.db = db
	return &manager, nil
}

// FindAll - returns all checks in the database
func (state ChecksManager) FindAll() ([]*Check, error) {
	checks := []*Check{}
	err := state.db.Find(&checks).Error
	return checks, err
}

// Find - returns a specific checks based on id in database
func (state ChecksManager) Find(id uint) (*Check, error) {
	check := &Check{}
	err := state.db.Where("id = ?", id).Find(check).Error
	return check, err
}

// FindWithClientAndCommand - finds all checks based on client id and command id, it also orders by timestamp in desc order
func (state ChecksManager) FindWithClientAndCommand(client uint, cmd uint) (*Check, error) {
	check := &Check{}
	err := state.db.Where("client_id = ? AND command_id = ?", client, cmd).Order("created_at desc").Select(&check).Error
	return check, err
}

// Save - saves a check in the database
func (state *ChecksManager) Save(check *Check) error {
	return state.db.Save(&check).Error
}

// Update - updates a specifc check (Wrapper for finding, editing and saving)
func (state ChecksManager) Update(id uint) error {
	// Find the correct check
	check, err := state.Find(id)
	if err != nil {
		return err
	}

	// Modify the check
	check.Checked = true

	// Save the check to the database
	return state.Save(check)
}

// Create inserts a new check into the database
func (state *ChecksManager) Create(check *Check) error {
	return state.db.Create(check).Error
}
