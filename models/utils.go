package models

import (
	"sync"
	"time"

	"github.com/keiwi/utils"
	"github.com/keiwi/utils/models"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2/bson"
)

func FindAllClients(conn *nats.Conn) ([]models.Client, error) {
	requestData := utils.FindOptions{
		Sort: utils.Sort{"created_at"},
	}
	data, err := bson.MarshalJSON(requestData)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling")
	}
	msg, err := conn.Request("clients.retrieve.find", data, time.Duration(viper.GetInt("nats_delay"))*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting")
	}

	var clients []models.Client
	err = bson.UnmarshalJSON(msg.Data, &clients)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling")
	}
	return clients, nil
}

func FindAllGroups(conn *nats.Conn) ([]models.Group, error) {
	requestData := utils.FindOptions{
		Sort: utils.Sort{"created_at"},
	}
	data, err := bson.MarshalJSON(requestData)
	if err != nil {
		return nil, err
	}
	msg, err := conn.Request("groups.retrieve.find", data, time.Duration(viper.GetInt("nats_delay"))*time.Second)
	if err != nil {
		return nil, err
	}

	var groups []models.Group
	err = bson.UnmarshalJSON(msg.Data, &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func FindAllCommands(conn *nats.Conn) ([]models.Command, error) {
	requestData := utils.FindOptions{
		Sort: utils.Sort{"created_at"},
	}
	data, err := bson.MarshalJSON(requestData)
	if err != nil {
		return nil, err
	}
	msg, err := conn.Request("commands.retrieve.find", data, time.Duration(viper.GetInt("nats_delay"))*time.Second)
	if err != nil {
		return nil, err
	}

	var commands []models.Command
	err = bson.UnmarshalJSON(msg.Data, &commands)
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func FindCheck(conn *nats.Conn, filter utils.Filter) ([]models.Check, error) {
	requestData := utils.FindOptions{
		Filter: filter,
		Sort:   utils.Sort{"-created_at"},
	}
	data, err := bson.MarshalJSON(requestData)
	if err != nil {
		return nil, err
	}
	msg, err := conn.Request("checks.retrieve.find", data, time.Duration(viper.GetInt("nats_delay"))*time.Second)
	if err != nil {
		return nil, err
	}

	var checks []models.Check
	err = bson.UnmarshalJSON(msg.Data, &checks)
	if err != nil {
		return nil, err
	}

	return checks, nil
}

func FindWithClientAndCommand(conn *nats.Conn, clientID, commandID bson.ObjectId) (*models.Check, error) {
	checks, err := FindCheck(conn, utils.Filter{"client_id": clientID, "command_id": commandID})
	if err != nil {
		return nil, err
	}

	if len(checks) <= 0 {
		return nil, errors.New("could not find any checks")
	}
	return &checks[0], nil
}

func FindCheckID(conn *nats.Conn, checkID bson.ObjectId) (*models.Check, error) {
	checks, err := FindCheck(conn, utils.Filter{"_id": checkID})
	if err != nil {
		return nil, err
	}

	if len(checks) <= 0 {
		return nil, errors.New("could not find any checks")
	}
	return &checks[0], nil
}

func UpdateCheck(conn *nats.Conn, checkID bson.ObjectId) (bool, error) {
	check, err := FindCheckID(conn, checkID)
	if err != nil || check == nil {
		return false, nil
	}

	requestData := utils.UpdateOptions{
		Filter:  utils.Filter{"_id": checkID},
		Updates: utils.Updates{"$set": bson.M{"checked": true}},
	}

	data, err := bson.MarshalJSON(requestData)
	if err != nil {
		return true, err
	}
	return true, conn.Publish("checks.update.send", data)
}

func CreateCheck(conn *nats.Conn, check *models.Check) error {
	data, err := bson.MarshalJSON(check)
	if err != nil {
		return err
	}
	return conn.Publish("checks.create.send", data)
}

// NewManager - Will read from database and create new virtual clients
func NewManager(conn *nats.Conn) (*Manager, error) {
	clients := &Manager{rw: new(sync.RWMutex)}

	// Gather all information from database, clients, groups, commands
	cls, err := FindAllClients(conn)
	if err != nil {
		return clients, errors.Wrap(err, "error finding all clients")
	}

	groups, err := FindAllGroups(conn)
	if err != nil {
		return clients, errors.Wrap(err, "error finding all groups")
	}

	cmds, err := FindAllCommands(conn)
	if err != nil {
		return clients, errors.Wrap(err, "error finding all commands")
	}

	// Create virtual commands groups
	c := ConvertCommands(cmds)
	g := ConvertGroups(groups, c)

	// Loop through all the clients
	for _, client := range cls {
		// Create a new virtual client
		cl := NewClient(&client)

		// Split the group string and loop through all of its groups
		for _, groupID := range client.GroupIDs {
			// loop through all the virtual groups
			for _, group := range g {
				// Check for the correct group and add it to the client
				if group.ID() == groupID {
					cl.AddGroup(group)

					// Loop through all commands to check for latest check
					for cmd := range group.IterCommands() {

						check, cerr := FindWithClientAndCommand(conn, cl.ID(), cmd.ID())
						if err != nil {
							return clients, errors.Wrap(cerr, "error finding check with client and command id")
						}

						ch := NewCheck(check, cmd)
						ch.SetGroup(group)
						cl.AddCheck(ch)
					}
				}
			}
		}
		clients.AddClient(cl)
	}
	return clients, nil
}

func ConvertCommands(cmds []models.Command) []*Command {
	c := make([]*Command, len(cmds))
	for i, cmd := range cmds {
		c[i] = &Command{rw: new(sync.RWMutex), id: cmd.ID, command: cmd.Command}
	}
	return c
}

func ConvertGroups(groups []models.Group, cmds []*Command) []*Group {
	g := make([]*Group, len(groups))

	for i, group := range groups {
		mGroup := &Group{
			rw:   new(sync.RWMutex),
			name: group.Name,
			id:   group.ID,
		}

		for _, cmd := range group.Commands {
			for _, c := range cmds {
				if cmd.CommandID == c.ID() {
					mCmd := c.Clone()
					mCmd.SetGroupID(group.ID)
					mCmd.SetInterval(cmd.NextCheck)
					mCmd.SetFailOnError(cmd.StopError)
					mGroup.AddCommand(mCmd)
					break
				}
			}
		}

		g[i] = mGroup
	}

	return g
}
