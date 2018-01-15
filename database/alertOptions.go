package database

import (
	"time"
)

// AlertOptions struct
type AlertOptions struct {
	ID        uint       `db:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
	ClientID  uint       `db:"client_id"`
	CommandID uint       `db:"command_id"`
	Alert     string     `db:"alert"`
	Value     string     `db:"value"`
	Count     uint       `db:"count"`
	Delay     uint       `db:"delay"`
	Service   string     `db:"service"`
}

// AlertOptionsManager struct
type AlertOptionsManager struct {
	db *DB
}

// NewAlertOptionsManager - Creates a new *AlertOptionsManager that can be used for managing alert options.
func NewAlertOptionsManager(db *DB) (*AlertOptionsManager, error) {
	db.AutoMigrate(&AlertOptions{})
	manager := AlertOptionsManager{}
	manager.db = db
	return &manager, nil
}

// FindAll - returns all alert options in the database
func (state AlertOptionsManager) FindAll() ([]AlertOptions, error) {
	alerts := []AlertOptions{}
	err := state.db.Find(&alerts).Error
	return alerts, err
}

// Find - returns a specific alert option based on id in database
func (state AlertOptionsManager) Find(id uint) (AlertOptions, error) {
	alert := AlertOptions{}
	err := state.db.Where("id = ?", id).Find(alert).Error
	return alert, err
}
