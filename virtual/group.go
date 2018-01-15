package virtual

import (
	"sync"
)

// NewGroup creates a new virtual group
func NewGroup(name string) *Group {
	return &Group{
		rw:       new(sync.RWMutex),
		Name:     name,
		Commands: []*Command{},
	}
}

// Group is the structure of a virtual group
type Group struct {
	rw       *sync.RWMutex
	Name     string
	Commands []*Command
}

// GetName returns the name of the group in a safe way
func (g Group) GetName() (name string) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.Name
}

// GetCommands returns all of the commands in a safe way
func (g Group) GetCommands() (cmds []*Command) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.Commands
}

// GetCommand tries to find a command in memory in a safe way, the index is based on the in memory array
func (g Group) GetCommand(i int) (cmd *Command) {
	if i >= g.Count() {
		return
	}
	g.rw.RLock()
	defer g.rw.RUnlock()
	return g.Commands[i]
}

// GetCommandByID tries to find a command in memory in a safe way, based on the command ID
func (g Group) GetCommandByID(id uint) (cmd *Command) {
	for c := range g.IterCommands() {
		if c.GetID() == id {
			cmd = c
			break
		}
	}
	return
}

// HasCommand will return if the command ID exists or not in memory
func (g Group) HasCommand(id uint) (ok bool) {
	c := g.GetCommandByID(id)
	if c != nil {
		return true
	}
	return false
}

// Count returns the amount of commands that exists in memory
func (g Group) Count() (count int) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return len(g.Commands)
}

// SetName modifies the group name
func (g *Group) SetName(name string) {
	g.rw.Lock()
	defer g.rw.Unlock()
	g.Name = name
}

// AddCommand add a new command in memory
func (g *Group) AddCommand(cmd *Command) {
	g.rw.Lock()
	defer g.rw.Unlock()
	g.Commands = append(g.Commands, cmd)
}

// RemoveCommand removes a command from memory, the index is based on the in memory array
func (g *Group) RemoveCommand(i int) (ok bool) {
	if i >= g.Count() || i < 0 {
		return false
	}
	g.rw.Lock()
	defer g.rw.Unlock()
	g.Commands = append(g.Commands[:i], g.Commands[i+1:]...)
	return true
}

// RemoveByID removes a command from memory based on the command id
func (g *Group) RemoveCommandByID(id uint) (ok bool) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for i := g.Count() - 1; i >= 0; i-- {
		c := g.GetCommand(i)
		if c != nil && c.GetID() == id {
			g.Commands = append(g.Commands[:i], g.Commands[i+1:]...)
			return true
		}
	}
	return false
}

// IterCommands - Will return a channel and loop through all the commands in a safe way and pass it to the channel
func (g Group) IterCommands() <-chan *Command {
	ch := make(chan *Command, g.Count())
	go func() {
		g.rw.RLock()
		defer g.rw.RUnlock()
		for _, cmd := range g.Commands {
			ch <- cmd
		}
		close(ch)
	}()
	return ch
}
