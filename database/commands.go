package database

import (
	"time"

	// import mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Command struct
type Command struct {
	ID          uint       `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time  `json:"-"`
	UpdatedAt   time.Time  `json:"-"`
	DeletedAt   *time.Time `sql:"index" json:"-"`
	Command     string     `gorm:"not null" json:"command"`
	Namn        string     `gorm:"not null;unique" json:"namn"`
	Description string     `gorm:"not null" json:"description"`
	Format      string     `json:"format"`
}

// CommandsManager struct
type CommandsManager struct {
	db *DB
}

// NewCommandsManager - Creates a new *CommandsManager that can be used for managing commands.
func NewCommandsManager(db *DB) (*CommandsManager, error) {
	db.AutoMigrate(&Command{})
	manager := CommandsManager{}
	manager.db = db
	return &manager, nil
}

// FindAll - returns all commands in database
func (state CommandsManager) FindAll() ([]*Command, error) {
	commands := []*Command{}
	err := state.db.Find(&commands).Error
	return commands, err
}

// Find - returns a specific command based in datbase
func (state CommandsManager) Find(id uint) (*Command, error) {
	command := &Command{}
	err := state.db.Where("id = ?", id).Find(command).Error
	return command, err
}
