package database

import (
	"time"
)

// Alert struct
type Alert struct {
	ID        uint       `db:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
	AlertID   uint       `db:"alert_id"`
	ClientID  uint       `db:"client_id"`
	Value     string     `db:"value"`
}

// AlertManager struct
type AlertManager struct {
	db *DB
}

// NewAlertManager - Creates a new *AlertManager that can be used for managing alert .
func NewAlertManager(db *DB) (*AlertManager, error) {
	db.AutoMigrate(&Alert{})
	manager := AlertManager{}
	manager.db = db
	return &manager, nil
}

// FindAll - returns all alerts in the database
func (state AlertManager) FindAll() ([]Alert, error) {
	alerts := []Alert{}
	err := state.db.Find(&alerts).Error
	return alerts, err
}

// Find - returns a specific alert based on id in database
func (state AlertManager) Find(id uint) (Alert, error) {
	alert := Alert{}
	err := state.db.Where("id = ?", id).Find(&alert).Error
	return alert, err
}

// FindWithAlertAndClient - returns a specifci alert based on alert and client id in the database
func (state AlertManager) FindWithAlertAndClient(alert uint, client uint) (*Alert, error) {
	a := &Alert{}
	err := state.db.Where("alert_id = ? AND client_id = ?", alert, client).Order("created_at desc").Select(&a).Error
	return a, err
}

// Save - saves an alert in the database
func (state *AlertManager) Save(alert *Alert) error {
	return state.db.Save(&alert).Error
}
