package models

import (
	"sync"

	"gopkg.in/mgo.v2/bson"
)

// Manager struct, contains all the virtual clients
type Manager struct {
	rw      *sync.RWMutex
	clients []*Client
}

// Client - Return a specific client, the index is based on in memory array
func (c Manager) Client(i int) (cl *Client) {
	if i >= c.Length() {
		return
	}

	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.clients[i]
}

// Clients - Return all the clients
func (c Manager) Clients() (cls []*Client) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.clients
}

// ClientByID - Check for a specifci client based on the client ID
func (c Manager) ClientByID(id bson.ObjectId) (cl *Client) {
	for cli := range c.IterClients() {
		if cli.ID() == id {
			cl = cli
			break
		}
	}
	return
}

// AddClient - Add a new client to in memory array
func (c *Manager) AddClient(client *Client) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.clients = append(c.clients, client)
}

// RemoveClient - Remove a client from memory, the index is based on in memory array
func (c *Manager) RemoveClient(i int) (ok bool) {
	if i >= c.Length() || i < 0 {
		return false
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.clients = append(c.clients[:i], c.clients[i+1:]...)
	return true
}

// RemoveClientByID - Remove a client from memory based on the id
func (c *Manager) RemoveClientByID(id bson.ObjectId) (ok bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := c.Length() - 1; i >= 0; i-- {
		cl := c.Client(i)
		if cl != nil && cl.ID() == id {
			c.clients = append(c.clients[:i], c.clients[i+1:]...)
			return true
		}
	}
	return false
}

// IterClients - Will return a channel and loop through all the clients in a safe way and pass it to the channel
func (c Manager) IterClients() <-chan *Client {
	ch := make(chan *Client, c.Length())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, cl := range c.clients {
			ch <- cl
		}
		close(ch)
	}()
	return ch
}

// Length - Will return the amount of clients in memory
func (c Manager) Length() (count int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.clients)
}
