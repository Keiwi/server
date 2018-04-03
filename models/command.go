package models

import (
	"sync"

	"gopkg.in/mgo.v2/bson"
)

// NewCommand creates a new command from parameters
func NewCommand(command string, commandID bson.ObjectId, interval int, failonerror bool) *Command {
	return &Command{
		rw:          new(sync.RWMutex),
		command:     command,
		id:          commandID,
		interval:    interval,
		failonerror: failonerror,
	}
}

// Command is the structure for virtual commands
type Command struct {
	rw          *sync.RWMutex
	command     string        // The command to send to the client
	id          bson.ObjectId // The ID in database
	groupid     bson.ObjectId // The ID on the group in the database
	interval    int           // The time between checks (in seconds)
	failonerror bool          // If the check should stop when it gets an error or not
}

// Command returns the command in a safe way
func (c Command) Command() (cmd string) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.command
}

// ID returns the commands ID In a safe way
func (c Command) ID() bson.ObjectId {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.id
}

// GroupID returns the group ID in a safe way
func (c Command) GroupID() bson.ObjectId {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.groupid
}

// Interval returns the next check value in a safe way
func (c Command) Interval() (next int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.interval
}

// FailOnError returns the stop error value in a safe way
func (c Command) FailOnError() (stop bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.failonerror
}

// SetGroupID modifies the group ID in a safe way
func (c *Command) SetGroupID(id bson.ObjectId) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.groupid = id
}

// SetCommand modifies the command in a safe way
func (c *Command) SetCommand(command string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.command = command
}

// SetInterval modifies the time between checks in a safe way
func (c *Command) SetInterval(next int) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.interval = next
}

// SetFailOnError modifies if the check should stop on an error or not in a safe way
func (c *Command) SetFailOnError(stop bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.failonerror = stop
}

// Clone copies all the values of a command and returns a new command in a safe way
func (c *Command) Clone() *Command {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return &Command{rw: new(sync.RWMutex), id: c.id, command: c.command, interval: c.interval, failonerror: c.failonerror}
}
