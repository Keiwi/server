package models

import (
	"sync"

	"gopkg.in/mgo.v2/bson"
)

// NewGroup creates a new virtual group
func NewGroup(name string) *Group {
	return &Group{
		rw:       new(sync.RWMutex),
		name:     name,
		commands: []*Command{},
	}
}

// Group is the structure of a virtual group
type Group struct {
	rw       *sync.RWMutex
	id       bson.ObjectId
	name     string
	commands []*Command
}

// ID returns the name of the group in a safe way
func (g Group) ID() bson.ObjectId {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.id
}

// Name returns the name of the group in a safe way
func (g Group) Name() (name string) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.name
}

// Commands returns all of the commands in a safe way
func (g Group) Commands() (cmds []*Command) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.commands
}

// Command tries to find a command in memory in a safe way, the index is based on the in memory array
func (g Group) Command(i int) (cmd *Command) {
	if i >= g.Length() {
		return
	}
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.commands[i]
}

// CommandByID tries to find a command in memory in a safe way, based on the command ID
func (g Group) CommandByID(id bson.ObjectId) (cmd *Command) {
	for c := range g.IterCommands() {
		if c.ID() == id {
			cmd = c
			break
		}
	}
	return
}

// HasCommand will return if the command ID exists or not in memory
func (g Group) HasCommand(id bson.ObjectId) (ok bool) {
	return g.CommandByID(id) != nil
}

// Length returns the amount of commands that exists in memory
func (g Group) Length() (count int) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return len(g.commands)
}

// SetName modifies the group name
func (g *Group) SetName(name string) {
	g.rw.Lock()
	defer g.rw.Unlock()
	g.name = name
}

// AddCommand add a new command in memory
func (g *Group) AddCommand(cmd *Command) {
	g.rw.Lock()
	defer g.rw.Unlock()
	g.commands = append(g.commands, cmd)
}

// RemoveCommand removes a command from memory, the index is based on the in memory array
func (g *Group) RemoveCommand(i int) (ok bool) {
	if i >= g.Length() || i < 0 {
		return false
	}
	g.rw.Lock()
	defer g.rw.Unlock()
	g.commands = append(g.commands[:i], g.commands[i+1:]...)
	return true
}

// RemoveByID removes a command from memory based on the command id
func (g *Group) RemoveCommandByID(id bson.ObjectId) (ok bool) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for i := g.Length() - 1; i >= 0; i-- {
		c := g.Command(i)
		if c != nil && c.ID() == id {
			g.commands = append(g.commands[:i], g.commands[i+1:]...)
			return true
		}
	}
	return false
}

// IterCommands - Will return a channel and loop through all the commands in a safe way and pass it to the channel
func (g Group) IterCommands() <-chan *Command {
	ch := make(chan *Command, g.Length())
	go func() {
		g.rw.RLock()
		defer g.rw.RUnlock()
		for _, cmd := range g.commands {
			ch <- cmd
		}
		close(ch)
	}()
	return ch
}
