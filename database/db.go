package database

import (
	"fmt"

	"github.com/jinzhu/gorm"

	// mysql db driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB abstraction
type DB struct {
	*gorm.DB
}

// Database struct
type Database struct {
	conn         *DB
	Alerts       *AlertManager
	AlertOptions *AlertOptionsManager
	Checks       *ChecksManager
	Clients      *ClientsManager
	Commands     *CommandsManager
	Groups       *GroupsManager
}

// NewMysqlDB - mysql database
func NewMysqlDB(user, password, host, port, dbname string) (*Database, error) {
	conn, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, dbname))
	if err != nil {
		return nil, err
	}

	if err = conn.DB().Ping(); err != nil {
		return nil, err
	}

	conn.LogMode(false)

	aDB := &DB{conn}
	db := &Database{conn: aDB}

	alertsmgr, err := NewAlertManager(db.conn)
	if err != nil {
		return nil, err
	}

	alertsoptsmgr, err := NewAlertOptionsManager(db.conn)
	if err != nil {
		return nil, err
	}

	checksmgr, err := NewChecksManager(db.conn)
	if err != nil {
		return nil, err
	}

	clientsmgr, err := NewClientsManager(db.conn)
	if err != nil {
		return nil, err
	}

	commandsmgr, err := NewCommandsManager(db.conn)
	if err != nil {
		return nil, err
	}

	groupsmgr, err := NewGroupsManager(db.conn)
	if err != nil {
		return nil, err
	}

	db.Alerts = alertsmgr
	db.AlertOptions = alertsoptsmgr
	db.Checks = checksmgr
	db.Clients = clientsmgr
	db.Commands = commandsmgr
	db.Groups = groupsmgr

	return db, nil
}

// Close - Closes all mysql connections
func (d *Database) Close() error {
	return d.conn.Close()
}
