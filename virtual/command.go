package virtual

import (
	"sync"
)

// NewCommand creates a new command from parameters
func NewCommand(command string, commandID uint, nextCheck int, stopError bool) *Command {
	return &Command{
		rw:        new(sync.RWMutex),
		Command:   command,
		ID:        commandID,
		NextCheck: nextCheck,
		StopError: stopError,
	}
}

// Command is the structure for virtual commands
type Command struct {
	rw        *sync.RWMutex
	Command   string // The command to send to the client
	ID        uint   // The ID in database
	GroupID   uint   // The ID on the group in the database
	NextCheck int    // The time between checks (in seconds)
	StopError bool   // If the check should stop when it gets an error or not
}

// GetCommand returns the command in a safe way
func (c Command) GetCommand() (cmd string) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Command
}

// GetID returns the commands ID In a safe way
func (c Command) GetID() (id uint) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.ID
}

// GetGroupID returns the group ID in a safe way
func (c Command) GetGroupID() uint {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.GroupID
}

// GetNextCheck returns the next check value in a safe way
func (c Command) GetNextCheck() (next int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.NextCheck
}

// GetStopError returns the stop error value in a safe way
func (c Command) GetStopError() (stop bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.StopError
}

// SetGroupID modifies the group ID in a safe way
func (c *Command) SetGroupID(id uint) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.GroupID = id
}

// SetCommand modifies the command in a safe way
func (c *Command) SetCommand(command string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Command = command
}

// SetNextCheck modifies the time between checks in a safe way
func (c *Command) SetNextCheck(next int) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.NextCheck = next
}

// SetStopError modifies if the check should stop on an error or not in a safe way
func (c *Command) SetStopError(stop bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.StopError = stop
}

// Clone copies all the values of a command and returns a new command in a safe way
func (c *Command) Clone() *Command {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return &Command{rw: new(sync.RWMutex), ID: c.ID, Command: c.Command, NextCheck: c.NextCheck, StopError: c.StopError}
}
