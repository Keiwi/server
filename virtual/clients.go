package virtual

import (
	"strings"
	"sync"

	"github.com/keiwi/server/database"
)

// Clients struct, contains all the virtual clients
type Clients struct {
	rw      *sync.RWMutex
	Clients []*Client
}

// GetClient - Return a specific client, the index is based on in memory array
func (c Clients) GetClient(i int) (cl *Client) {
	if i >= c.Count() {
		return
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Clients[i]
}

// GetClients - Return all the clients
func (c Clients) GetClients() (cls []*Client) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Clients
}

// GetClientByID - Check for a specifci client based on the client ID
func (c Clients) GetClientByID(id uint) (cl *Client) {
	for cli := range c.IterClients() {
		if cli.GetID() == id {
			cl = cli
			break
		}
	}
	return
}

// AddClient - Add a new client to in memory array
func (c *Clients) AddClient(client *Client) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Clients = append(c.Clients, client)
}

// RemoveClient - Remove a client from memory, the index is based on in memory array
func (c *Clients) RemoveClient(i int) (ok bool) {
	if i >= c.Count() || i < 0 {
		return false
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Clients = append(c.Clients[:i], c.Clients[i+1:]...)
	return true
}

// RemoveClientByID - Remove a client from memory based on the id
func (c *Clients) RemoveClientByID(id uint) (ok bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := c.Count() - 1; i >= 0; i-- {
		cl := c.GetClient(i)
		if cl != nil && cl.GetID() == id {
			c.Clients = append(c.Clients[:i], c.Clients[i+1:]...)
			return true
		}
	}
	return false
}

// IterClients - Will return a channel and loop through all the clients in a safe way and pass it to the channel
func (c Clients) IterClients() <-chan *Client {
	ch := make(chan *Client, c.Count())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, cl := range c.Clients {
			ch <- cl
		}
		close(ch)
	}()
	return ch
}

// Count - Will return the amount of clients in memory
func (c Clients) Count() (count int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.Clients)
}

// NewClients - Will read from database and create new virtual clients
func NewClients(db *database.Database) (*Clients, error) {
	clients := &Clients{rw: new(sync.RWMutex)}

	// Gather all information from database, clients, groups, commands
	cls, err := db.Clients.FindAll()
	if err != nil {
		return clients, err
	}

	groups, err := db.Groups.FindAll()
	if err != nil {
		return clients, err
	}

	cmds, err := db.Commands.FindAll()
	if err != nil {
		return clients, err
	}

	// Create virtual commands groups
	c := GetVirtualCommands(cmds)
	g := GetVirtualGroups(groups, c)

	// Loop through all the clients
	for _, client := range cls {
		// Create a new virtual client
		cl := NewClient(client)

		// Split the group string and loop through all of its groups
		for _, group := range strings.Split(client.GroupNames, ",") {
			// loop through all the virtual groups
			for _, gg := range g {
				// Check for the correct group and add it to the client
				if gg.GetName() == group {
					cl.AddGroup(gg)

					// Loop through all commands to check for latest check
					for cmd := range gg.IterCommands() {
						check, cerr := db.Checks.FindWithClientAndCommand(cl.GetID(), cmd.GetID())
						if err != nil {
							return clients, cerr
						}

						ch := NewCheck(check, cmd)
						ch.SetGroup(gg)
						cl.AddCheck(ch)
					}
				}
			}
		}
		clients.AddClient(cl)
	}
	return clients, err
}
