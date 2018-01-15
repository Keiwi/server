package http

import (
	"github.com/keiwi/server/virtual"
	"github.com/keiwi/utils"
)

// GroupHandler is the struct that will handle all the API Requests regarding groups
type GroupHandler struct{}

// Insert will be ran when someone add a group to a client
func (GroupHandler) Insert(r *Handler) {
	if r.Request.GroupName == "" {
		r.OutputMessage(true, "You need to supply a group name")
		return
	}

	// Get the client from database
	cl := r.GetServer().GetClients().GetClientByID(r.Request.ID)
	if cl == nil {
		r.OutputMessage(true, "Can't find a client with this ID")
		return
	}

	// Check if client already is in this group
	if cl.HasGroupByName(r.Request.GroupName) {
		r.OutputMessage(true, "Client already belongs to this group in cache")
		return
	}

	// Get the group from the database
	groups, err := virtual.GetGroupFromName(r.GetServer().GetDatabase(), r.Request.GroupName)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.OutputMessage(true, "Internal error")
		return
	}

	// Check if the group actually exists in the database
	if groups == nil || len(groups) <= 0 {
		r.OutputMessage(true, "Can't find the group in database.")
		return
	}

	// Loop through all the groups commands
	for cmd := range groups[0].IterCommands() {
		// Find the latest checks
		check, err := r.GetServer().GetDatabase().Checks.FindWithClientAndCommand(cl.GetID(), cmd.GetID())
		if err != nil {
			utils.Log.WithField("error", err).Error("Internal error")
			r.OutputMessage(true, "Internal error")
			return
		}

		// Create a new virtual check and add it to the client
		ch := virtual.NewCheck(check, cmd)
		cl.AddCheck(ch)
	}
	cl.AddGroup(groups[0])
	r.OutputMessage(false, "Added the group to the client")
}

// Update is ran when updating a group
func (GroupHandler) Update(r *Handler) {
	group, err := r.GetServer().GetDatabase().Groups.Find(r.Request.GroupID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.OutputMessage(true, "Internal error")
		return
	}

	// loop through all references of the group and update with the new settings
	for cl := range r.GetServer().GetClients().IterClients() {
		for g := range cl.IterGroups() {
			for c := range g.IterCommands() {
				if c.GetGroupID() == group.ID {
					c.SetNextCheck(group.NextCheck)
					c.SetStopError(group.StopError)
				}
			}
		}
	}
	r.OutputMessage(false, "Updated the group in cache")
}

// Delete will be ran when deleting a group
func (GroupHandler) Delete(r *Handler) {
	if r.Request.GroupName == "" {
		r.OutputMessage(true, "You need to supply a group name.")
		return
	}

	//if r.Request.ID != -1 {
	cl := r.GetServer().GetClients().GetClientByID(r.Request.ID)
	if cl == nil {
		r.OutputMessage(true, "Can't find a client with this ID")
		return
	}

	ok := cl.RemoveGroupsByName(r.Request.GroupName)
	if !ok {
		r.OutputMessage(true, "Client does not belong to this group in cache")
	} else {
		r.OutputMessage(false, "Group has now been removed from the client in cache")
	}
	return
	/* } else {
		for cl := range r.Server.GetClients().IterClients() {
			cl.RemoveGroupsByName(r.Request.GroupName)
		}
		r.OutputMessage(false, "Group has now been removed from all clients in cache")
	} */
}
