package virtual

import (
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/keiwi/server/database"
	"github.com/keiwi/utils"

	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"strconv"
)

// NewClient - Creates a new virtual client
func NewClient(cl *database.Client) *Client {
	return &Client{
		rw: new(sync.RWMutex),
		IP: cl.IP,
		ID: cl.ID,
	}
}

// Client is the virtual client struct
type Client struct {
	rw         *sync.RWMutex
	IP         string
	ID         uint
	Groups     []*Group
	Checks     []*Check
	groupNames []string
}

// GetIP returns the clients ip
func (c Client) GetIP() (ip string) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.IP
}

// GetID returns the clients database ID
func (c Client) GetID() (id uint) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.ID
}

// GetGroup returns a specific group that the client belongs to, the index is based on the array in memory
func (c Client) GetGroup(i int) (group *Group) {
	if i >= c.CountGroups() {
		return
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Groups[i]
}

// GetGroups returns all virtual groups that the client belongs to
func (c Client) GetGroups() (groups []*Group) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Groups
}

// HasGroupByName checks whether or not the client belongs to a specific group
func (c Client) HasGroupByName(name string) bool {
	for g := range c.IterGroups() {
		if g.GetName() == name {
			return true
		}
	}
	return false
}

// GetGroupsByName returns all virtual groups that the client belongs to based on a name
func (c Client) GetGroupsByName(name string) (groups []*Group) {
	groups = []*Group{}
	for g := range c.IterGroups() {
		if g.GetName() == name {
			groups = append(groups, g)
		}
	}
	return
}

// GetGroupsByCommand returns all virtual groups that the client belongs to based on a specific command id
func (c Client) GetGroupsByCommand(commandID uint) (groups []*Group) {
	groups = []*Group{}
	for g := range c.IterGroups() {
		for cmd := range g.IterCommands() {
			if cmd.GetID() == commandID {
				groups = append(groups, g)
				break
			}
		}
	}
	return
}

// GetCheck returns a specific check that belongs to the client, the index is based on the array in memory
func (c Client) GetCheck(i int) (check *Check) {
	if i >= c.CountChecks() {
		return
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Checks[i]
}

// GetChecks returns all checks that belongs to the client
func (c Client) GetChecks() (checks []*Check) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Checks
}

// GetCheckByPastID returns a specific check that belongs to the client based on the check ID
func (c Client) GetCheckByPastID(id uint) (check *Check) {
	for che := range c.IterChecks() {
		if che.GetID() == id {
			return che
		}
	}
	return nil
}

// GetChecksByCommandID returns a specific check that belongs to the client based on the command ID
func (c Client) GetChecksByCommandID(id uint) []*Check {
	var checks []*Check
	for ch := range c.IterChecks() {
		if ch.GetID() == id {
			checks = append(checks, ch)
		}
	}
	return checks
}

// SetIP modifies the client IP
func (c Client) SetIP(ip string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.IP = ip
}

// IterGroups will return a channel and loop through all the groups in a safe way and pass it to the channel
func (c Client) IterGroups() <-chan *Group {
	ch := make(chan *Group, c.CountGroups())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, group := range c.Groups {
			ch <- group
		}
		close(ch)
	}()
	return ch
}

// IterChecks will return a channel and loop through all the checks in a safe way and pass it to the channel
func (c Client) IterChecks() <-chan *Check {
	ch := make(chan *Check, c.CountChecks())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, check := range c.Checks {
			ch <- check
		}
		close(ch)
	}()
	return ch
}

// CountGroups will return the amount of groups that the client belongs to
func (c Client) CountGroups() (count int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.Groups)
}

// CountChecks will return the amount of checks that belongs to the client
func (c Client) CountChecks() (count int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.Checks)
}

// AddGroup will add a group to the client
func (c *Client) AddGroup(group *Group) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.groupNames = append(c.groupNames, group.GetName())
	c.Groups = append(c.Groups, group)
}

// AddCheck will add a check to the client
func (c *Client) AddCheck(check *Check) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Checks = append(c.Checks, check)
}

// RemoveGroup will remove a group from the client, the index is based on the array in memory
func (c *Client) RemoveGroup(i int) (ok bool) {
	if i >= c.CountGroups() || i < 0 {
		return false
	}
	c.rw.Lock()
	group := c.Groups[i]
	c.Groups = append(c.Groups[:i], c.Groups[i+1:]...)
	c.rw.Unlock()
	for cmd := range group.IterCommands() {
		c.RemoveCheckByGroupID(cmd.GetGroupID())
	}
	return
}

// RemoveGroupsByName will remove a group from the client based on the groups name
func (c *Client) RemoveGroupsByName(name string) (ok bool) {
	var checks []uint
	c.rw.Lock()
	for i := len(c.Groups) - 1; i >= 0; i-- {
		g := c.Groups[i]
		if g != nil && g.GetName() == name {
			c.Groups = append(c.Groups[:i], c.Groups[i+1:]...)
			for cmd := range g.IterCommands() {
				checks = append(checks, cmd.GetGroupID())
			}
		}
	}
	c.rw.Unlock()
	if len(checks) > 0 {
		for _, i := range checks {
			c.RemoveCheckByGroupID(i)
		}
		return true
	}
	return false
}

// RemoveGroupsByCommand will remove a group from the client based on the command id
func (c *Client) RemoveGroupsByCommand(commandID uint) (ok bool) {
	var check uint
	var found bool

	c.rw.Lock()                               // Lock read and write
	for i := len(c.Groups) - 1; i >= 0; i-- { // Loop all of the groups
		g := c.Groups[i]                    // Grab the group
		for cmd := range g.IterCommands() { // Loop all of the commands
			if cmd.GetID() == commandID { // Check if it's the correct command
				c.Groups = append(c.Groups[:i], c.Groups[i+1:]...) // Remove the group from the list
				check = cmd.GetGroupID()                           // Add the check as we need to remove the check from the client when the user isn't in the group anymore
				found = true                                       // Set found to true so that we know we should remove the checks
				break
			}
		}
	}
	c.rw.Unlock() // Unlock read and write

	if found {
		c.RemoveCheckByGroupID(check) // Remove the checks from client
		return true
	}
	return false
}

// RemoveCheck will remove a specific check from the client, the index is based on the array in memory
func (c *Client) RemoveCheck(i int) (ok bool) {
	if i >= c.CountChecks() || i < 0 {
		return false
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Checks = append(c.Checks[:i], c.Checks[i+1:]...)
	return true
}

// RemoveCheckByID will remove a specific check from the client based on the check ID
func (c *Client) RemoveCheckByID(id uint) (ok bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := len(c.Checks) - 1; i >= 0; i-- {
		ch := c.Checks[i]
		if ch.GetID() == id {
			c.Checks = append(c.Checks[:i], c.Checks[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveCheckByGroupID will remove a specific check from the client based on the group id
func (c *Client) RemoveCheckByGroupID(id uint) (ok bool) {
	ok = false
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := len(c.Checks) - 1; i >= 0; i-- {
		ch := c.Checks[i]
		if ch.GetCommand().GetGroupID() == id {
			c.Checks = append(c.Checks[:i], c.Checks[i+1:]...)
			ok = true
		}
	}
	return ok
}

// SendMessage will create a TCP connection to the client then send a message to it and return the reply or error
func (c Client) SendMessage(message string) (resp string, err error) {
	// create connection to client
	conn, err := net.Dial("tcp", c.GetIP()+":3333")
	if err != nil {
		return
	}

	defer func() {
		err = conn.Close()
	}()

	// Send the message to the client
	if _, err = conn.Write([]byte(message)); err != nil {
		return
	}

	// Read the reply
	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	if err != nil {
		return
	}

	resp = string(reply[:n])
	err = nil
	return
}

var re = regexp.MustCompile("-port=\"?([\\d,-]+)\"?") // special case check, when checking for port when pinging a server

// SendCheck will send a check to the client and then save the check to the database
func (c *Client) SendCheck(db *database.Database, check *Check) string {
	// Start a check
	check.SetChecked(true)
	command := check.GetCommand()
	if err := db.Checks.Update(check.GetID()); err != nil {
		if err.Error() == "record not found" {
			ch := &database.Check{
				ID:        check.GetID(),
				CommandID: check.GetCommand().GetID(),
				ClientID:  c.GetID(),
				Response:  "",
				Checked:   check.GetChecked(),
				Error:     check.GetError(),
				Finished:  check.GetFinished(),
			}
			if err = db.Checks.Create(ch); err != nil {
				utils.Log.WithField("error", err).Error("error updating last check")
				return ""
			}
		} else {
			utils.Log.WithField("error", err).Error("error updating last check")
			return ""
		}
	}

	var resp string
	var err error

	// Special case when pinging as it should check if client is up and not send a command to the client
	if strings.HasPrefix(command.GetCommand(), "ping") {
		// Check if port is provided in the command if not default port it 3333
		ports := "3333"
		p := re.FindStringSubmatch(command.GetCommand())
		if len(p) >= 2 {
			ports = p[1]
		}

		// Ping the server to check if its on or not
		r, err := c.Ping(ports)
		b, _ := json.Marshal(r)
		e := ""
		if err != nil {
			e = err.Error()
		}
		resp = fmt.Sprintf(`{"error":"%s","ports":%s}`, e, b)
	} else {
		// Send the command to the client and wait for a reply
		resp, err = c.SendMessage(command.GetCommand())
	}

	// Check if there was an error in the connection or if the reply contains an error message
	if err != nil || !strings.Contains(resp, `"error":""`) {
		// An error occured so put the check error to true and if it was a connection issue, set the response to the error message
		check.SetError(true)
		if err != nil {
			resp = err.Error()
		}
	} else {
		check.SetError(false)
	}

	// Save the check to the database
	c.SaveCheck(db, check, resp)
	return resp
}

// SaveCheck will save a check to the database
func (c *Client) SaveCheck(db *database.Database, check *Check, resp string) {
	defer check.SetChecked(false)

	command := check.GetCommand()

	// Create a database check in memory
	ch := &database.Check{
		CommandID: command.GetID(),
		ClientID:  c.GetID(),
		Response:  resp,
		Error:     check.GetError(),
		Finished:  true,
	}

	// Save the new check to the datbase
	err := db.Checks.Create(ch)
	if err != nil {
		utils.Log.WithField("error", err).Error("error inserting new check")
		check.SetError(true)
		return
	}
	check.SetID(ch.ID)

	createdAt := ch.CreatedAt.Add(time.Duration(command.GetNextCheck()) * time.Second)
	check.SetTimestamp(createdAt)

	if !check.GetError() && command.GetStopError() {
		c.ResetCheck(check.GetGroup().GetName())
	}

	for a := range check.IterAlerts() {
		a.Check(resp, db)
	}
}

// ResetCheck will set error and checked to false on all checks with a specific name for the client
func (c *Client) ResetCheck(name string) {
	for ch := range c.IterChecks() {
		if ch.GetGroup().GetName() == name {
			ch.SetError(false)
			ch.SetChecked(false)
		}
	}
}

// Check loop through all clients check and check if it's time to do any checks.
func (c *Client) Check(db *database.Database) {
	// Start the loop
	for check := range c.IterChecks() {
		// If the check has already been started
		if check.GetChecked() {
			continue
		}

		// If the check contains any previous errors and the command has the condition to continue on error or not
		if check.GetError() && !check.GetCommand().GetStopError() {
			continue
		}

		// If its time for the next check or not
		t := check.GetTimestamp()
		if !t.IsZero() && time.Now().Before(t) {
			continue
		}

		// Send some debug message
		utils.Log.WithFields(log.Fields{
			"CommandID": check.GetCommand().GetID(),
			"ClientID":  c.GetID(),
		}).Info("Starting a check for client")

		// Start a check
		go c.SendCheck(db, check)
	}
}

// Default minimum and maximum port range
const (
	minTCPPort = 0
	maxTCPPort = 65535
)

// PingResult is the struct that will hold the result of a ping check
type PingResult struct {
	Port   uint16
	Result bool
}

// Port will check if a port is open on a server
func (c Client) Ping(port string) ([]PingResult, error) {
	var pings []PingResult
	var ports [][]uint16

	// split the port string with comma to check for multiple ports
	p := strings.Split(strings.Replace(port, " ", "", -1), ",")
	for _, port := range p {
		// Split he port on - in case its a port range
		pSplit := strings.Split(port, "-")

		// Convert the range to uint
		minPort, err := strconv.ParseUint(pSplit[0], 10, 16)
		// Range is not a valid integer
		if err != nil {
			return nil, fmt.Errorf("the value: %s can't be converted to an integer", pSplit[0])
		}
		// Check if minPort is within the normal port range (0-65535)
		if minPort < minTCPPort || minPort > maxTCPPort {
			return nil, fmt.Errorf("the value: %s is smaller or larger then the maximum TCP port range", pSplit[0])
		}
		maxPort := minPort

		// Check if it's a port range or not
		if len(pSplit) == 2 {
			// If it was a port range, convert the maximum port like above
			m, err := strconv.ParseUint(pSplit[1], 10, 16)
			if err != nil {
				return nil, fmt.Errorf("the value: %s can't be converted to an integer", pSplit[1])
			}
			if m < minTCPPort || m > maxTCPPort {
				return nil, fmt.Errorf("the value: %s is smaller or larger then the maximum TCP port range", pSplit[1])
			}
			maxPort = m
		}

		// Minimum port needs to be lower then maximum port
		if minPort > maxPort {
			return nil, fmt.Errorf("min value \"%d\" is larger then the max value \"%d\"", minPort, maxPort)
		}

		// Port is converted to integer and is now added to the port array
		ports = append(ports, []uint16{uint16(minPort), uint16(maxPort)})
	}

	var err error = nil
	// First loop through all the port ranges
	for _, port := range ports {
		// Second loop through all numbers between minPort and maxPort, will work if it was just a port range or simply a number
		for i := port[0]; i <= port[1]; i++ {
			// Start a TCP connection to aspeicifc port
			conn, e := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.GetIP(), i), time.Duration(1)*time.Second)
			// Check the result, if it contains an error, then the port is closed or the host is dead
			if e != nil {
				pings = append(pings, PingResult{Port: i, Result: false})
				err = fmt.Errorf("one or more servers failed")
			} else {
				conn.Close()
				pings = append(pings, PingResult{Port: i, Result: true})
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
	return pings, err
}
