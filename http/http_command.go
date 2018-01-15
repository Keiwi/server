package http

import (
	"github.com/keiwi/server/virtual"
	"github.com/keiwi/utils"
)

// CommandHandler is the struct for handling any API request regarding commands
type CommandHandler struct{}

// Insert will be ran when someone creates a new Command
func (CommandHandler) Insert(r *Handler) {
	// Check if GroupName was provided
	if r.Request.GroupName == "" {
		r.OutputMessage(true, "You need to supply a group name")
		return
	}

	// Find the command in the database
	cmd, err := r.GetServer().GetDatabase().Commands.Find(r.Request.CommandID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.OutputMessage(true, "Internal error")
		return
	}

	// Find the group in the database
	group, err := r.GetServer().GetDatabase().Groups.FindWithCommand(r.Request.GroupName, r.Request.CommandID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.OutputMessage(true, "Internal error")
		return
	}

	// Create a new virtual command
	command := &virtual.Command{
		Command:   cmd.Command,
		ID:        cmd.ID,
		GroupID:   group.ID,
		NextCheck: group.NextCheck,
		StopError: group.StopError,
	}

	// Loop through all clients
	for c := range r.GetServer().GetClients().IterClients() {
		// Check for the correct group
		getCheck := false
		for g := range c.IterGroups() {
			if g.GetName() == r.Request.GroupName {
				g.AddCommand(command)
				getCheck = true
			}
		}

		// Correct group found, now try to find latest check from database
		if getCheck {
			check, err := r.GetServer().GetDatabase().Checks.FindWithClientAndCommand(c.GetID(), command.GetID())
			if err != nil {
				utils.Log.WithField("error", err).Error("Internal error")
				r.OutputMessage(true, "Internal error")
				return
			}

			// Create a new virtual check from the database check
			ch := virtual.NewCheck(check, command)
			if err != nil {
				utils.Log.WithField("error", err).Error("Internal error")
				r.OutputMessage(true, "Internal error")
				return
			}

			// Add the virtual check to the client
			c.AddCheck(ch)
		}
	}
	r.OutputMessage(true, "Added the new command to the group in cache")
}

// Update is ran when someone tries to update a Group
func (CommandHandler) Update(r *Handler) {
	// Find the update in the database
	command, err := r.GetServer().GetDatabase().Commands.Find(r.Request.CommandID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.OutputMessage(true, "Internal error")
		return
	}

	// Update all the references to this command with the new changes
	for c := range r.GetServer().GetClients().IterClients() {
		for g := range c.IterGroups() {
			for cmd := range g.IterCommands() {
				if cmd.GetID() == command.ID {
					cmd.SetCommand(command.Command)
					break
				}
			}
		}
	}

	r.OutputMessage(true, "Successfully updated the command for all clients")
}

// Delete is ran when someone removes a command
func (CommandHandler) Delete(r *Handler) {
	// Remove all references of this command in memory
	for c := range r.GetServer().GetClients().IterClients() {
		for g := range c.IterGroups() {
			if r.Request.GroupName != "" {
				if r.Request.GroupName == g.GetName() {
					g.RemoveCommandByID(r.Request.CommandID)
				}
			} else {
				g.RemoveCommandByID(r.Request.CommandID)
			}
		}
	}
	r.OutputMessage(true, "Successfully removed the command.")
}
