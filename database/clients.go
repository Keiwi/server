package database

import (
	"time"

	// import mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Client struct
type Client struct {
	ID         uint       `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time  `json:"-"`
	UpdatedAt  time.Time  `json:"-"`
	DeletedAt  *time.Time `sql:"index" json:"-"`
	GroupNames string     `gorm:"not null" json:"group_names"`
	IP         string     `gorm:"not null" json:"ip"`
	Namn       string     `gorm:"not null;unique" json:"namn"`
}

// ClientsManager struct
type ClientsManager struct {
	db *DB
}

// NewClientsManager - Creates a new *ClientsManager that can be used for managing clients.
func NewClientsManager(db *DB) (*ClientsManager, error) {
	db.AutoMigrate(&Client{})
	manager := ClientsManager{}
	manager.db = db
	return &manager, nil
}

// FindAll - returns all clients in database
func (state ClientsManager) FindAll() ([]*Client, error) {
	clients := []*Client{}
	err := state.db.Find(&clients).Error
	return clients, err
}

// Find - returns a specific client based in datbase
func (state ClientsManager) Find(id uint) (*Client, error) {
	client := &Client{}
	err := state.db.Where("id = ?", id).Find(client).Error
	return client, err
}
