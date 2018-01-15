package virtual

import (
	"sync"

	"github.com/keiwi/server/database"
)

// GetVirtualGroups will loop through groups and commands and create "virtual" groups that will be used in the server
func GetVirtualGroups(groups []*database.Group, cmds []Command) []*Group {
	var g []*Group // Return value

	for _, group := range groups { // Loop through the provided groups
		var mGroup *Group = nil // Group in the models package

		for _, vGroup := range g { // Loop through the already existing virtual groups
			if vGroup.GetName() == group.GroupName { // Check if we have already created the virtual group or not
				mGroup = vGroup
				break
			}
		}

		if mGroup == nil { // Virtual group haven't been created
			mGroup = &Group{rw: new(sync.RWMutex), Name: group.GroupName}
			g = append(g, mGroup)
		}

		// Now add the commands to the virtual groups
		for _, c := range cmds {
			if c.GetID() == group.CommandID {
				cmd := c.Clone() // Create a new command (clone existing)
				cmd.SetGroupID(group.ID)
				cmd.SetNextCheck(group.NextCheck)
				cmd.SetStopError(group.StopError)
				mGroup.AddCommand(cmd)
				break
			}
		}
	}
	return g
}

// GetVirtualCommands will loop through commnads from the database and create Virtual commands
func GetVirtualCommands(cmds []*database.Command) []Command {
	var c []Command
	for _, cmd := range cmds {
		c = append(c, Command{rw: new(sync.RWMutex), ID: cmd.ID, Command: cmd.Command})
	}
	return c
}

// GetGroupFromName is a wrapper where it will get a group and it's commands from the database and then create a Virtual group
func GetGroupFromName(db *database.Database, name string) ([]*Group, error) {
	g, err := db.Groups.FindWithName(name)
	if err != nil || len(g) <= 0 {
		return nil, err
	}

	var c []*database.Command
	for _, group := range g {
		cmd, findErr := db.Commands.Find(group.CommandID)
		if findErr != nil {
			return nil, findErr
		}
		c = append(c, cmd)
	}

	cmds := GetVirtualCommands(c)
	groups := GetVirtualGroups(g, cmds)
	return groups, err
}
