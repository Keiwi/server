package http

import (
	"github.com/keiwi/server/virtual"
	"github.com/keiwi/utils"
)

// ClientHandler is a struct for handling the API for clients
type ClientHandler struct{}

// Insert is ran when someone tries to create a new client on the API
func (ClientHandler) Insert(r *Handler) {
	r.OutputMessage(true, "Unsupported method, please refer to the documentation")
}

// Update is ran when someone tries to update some data on a client
func (ClientHandler) Update(r *Handler) {
	// Find the client in the database
	c, err := r.GetServer().GetDatabase().Clients.Find(r.Request.ID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.OutputMessage(true, "Internal error")
		return
	}

	// Find the client in memory
	cl := r.GetServer().GetClients().GetClientByID(r.Request.ID)

	// If client does not exists in memory then it was newly created, if so add to memory
	if cl == nil {
		client := virtual.NewClient(c)
		r.GetServer().GetClients().AddClient(client)
		r.OutputMessage(false, "Client added")
	} else {
		cl.SetIP(c.IP)
		r.OutputMessage(false, "Client updated")
	}
}

// Delete is ran when someone tries to delete a client
func (ClientHandler) Delete(r *Handler) {
	b := r.GetServer().GetClients().RemoveClientByID(r.Request.ID)
	if !b {
		r.OutputMessage(true, "Client ID does not exists in cache")
	} else {
		r.OutputMessage(false, "Removed the client from cache")
	}
}
