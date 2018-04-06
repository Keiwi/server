package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/keiwi/utils/log"
	"github.com/keiwi/utils/models"
	"github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2/bson"
)

// NewClient - Creates a new virtual client
func NewClient(cl *models.Client) *Client {
	return &Client{
		rw: new(sync.RWMutex),
		ip: cl.IP,
		id: cl.ID,
	}
}

type Client struct {
	rw     *sync.RWMutex
	ip     string
	id     bson.ObjectId
	groups []*Group
	checks []*Check
	conn   net.Conn
}

// IP returns the clients IP
func (c Client) IP() (ip string) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.ip
}

// ID returns the clients database ID
func (c Client) ID() (id bson.ObjectId) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.id
}

// ID returns the clients database ID
func (c Client) Conn() (conn net.Conn) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.conn
}

// Group returns a specific group that the client belongs to, the index is based on the array index
func (c Client) Group(i int) (group *Group) {
	if i >= c.GroupsLength() {
		return nil
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.groups[i]
}

// Groups returns all virtual groups that the client belongs to
func (c Client) Groups() (groups []*Group) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.groups
}

// HasGroupByName returns if the client belongs to a group with a specific name
func (c Client) HasGroupByName(name string) bool {
	return c.GroupByName(name) != nil
}

// GroupByName returns the group with a specific name if the client belongs to it
func (c Client) GroupByName(name string) (group *Group) {
	for g := range c.IterGroups() {
		if g.Name() == name {
			return g
		}
	}
	return nil
}

// GroupByCommand will return all groups with a specific command ID
func (c Client) GroupByCommand(commandID bson.ObjectId) (groups []*Group) {
	groups = []*Group{}
	for g := range c.IterGroups() {
		for cmd := range g.IterCommands() {
			if cmd.ID() == commandID {
				groups = append(groups, g)
				break
			}
		}
	}
	return
}

// Check returns the specific check that belongs to the client, the index is based on the array index
func (c Client) Check(i int) (check *Check) {
	if i >= c.ChecksLength() {
		return nil
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.checks[i]
}

// Checks returns all checks that belongs to the client
func (c Client) Checks() (checks []*Check) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.checks
}

// CheckByPastID returns a specific check that belongs to the client based on the check ID
func (c Client) CheckByPastID(id bson.ObjectId) (check *Check) {
	for che := range c.IterChecks() {
		if che.ID() == id {
			return che
		}
	}
	return nil
}

// ChecksByCommandID returns a specific check that belongs to the client based on the command ID
func (c Client) ChecksByCommandID(id bson.ObjectId) []*Check {
	var checks []*Check
	for ch := range c.IterChecks() {
		if ch.ID() == id {
			checks = append(checks, ch)
		}
	}
	return checks
}

// SetIP modifies the client IP
func (c *Client) SetIP(ip string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.ip = ip
}

// SetIP modifies the client IP
func (c *Client) SetConn(conn net.Conn) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.conn = conn
}

// IterGroups will return a channel and loop through all the groups in a safe way and pass it to the channel
func (c Client) IterGroups() <-chan *Group {
	ch := make(chan *Group, c.GroupsLength())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, group := range c.groups {
			ch <- group
		}
		close(ch)
	}()
	return ch
}

// IterChecks will return a channel and loop through all the checks in a safe way and pass it to the channel
func (c Client) IterChecks() <-chan *Check {
	ch := make(chan *Check, c.ChecksLength())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, check := range c.checks {
			ch <- check
		}
		close(ch)
	}()
	return ch
}

// GroupsLength will return the amount of groups that the client belongs to
func (c Client) GroupsLength() (count int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.groups)
}

// ChecksLength will return the amount of checks that belongs to the client
func (c Client) ChecksLength() (count int) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.checks)
}

// AddGroup will add a group to the client
func (c *Client) AddGroup(group *Group) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.groups = append(c.groups, group)
}

// AddCheck will add a check to the client
func (c *Client) AddCheck(check *Check) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.checks = append(c.checks, check)
}

// RemoveGroup will remove a group from the client, the index is based on the array in memory
func (c *Client) RemoveGroup(i int) (ok bool) {
	if i >= c.GroupsLength() || i < 0 {
		return false
	}
	c.rw.Lock()
	group := c.groups[i]
	c.groups = append(c.groups[:i], c.groups[i+1:]...)
	c.rw.Unlock()
	for cmd := range group.IterCommands() {
		c.RemoveCheckByGroupID(cmd.GroupID())
	}
	return
}

// RemoveGroupsByName will remove a group from the client based on the groups name
func (c *Client) RemoveGroupsByName(name string) (ok bool) {
	var checks []bson.ObjectId
	c.rw.Lock()
	for i := len(c.groups) - 1; i >= 0; i-- {
		g := c.groups[i]
		if g != nil && g.Name() == name {
			c.groups = append(c.groups[:i], c.groups[i+1:]...)
			for cmd := range g.IterCommands() {
				checks = append(checks, cmd.GroupID())
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
func (c *Client) RemoveGroupsByCommand(commandID bson.ObjectId) (ok bool) {
	var check bson.ObjectId
	var found bool

	c.rw.Lock()                               // Lock read and write
	for i := len(c.groups) - 1; i >= 0; i-- { // Loop all of the groups
		g := c.groups[i]                    // Grab the group
		for cmd := range g.IterCommands() { // Loop all of the commands
			if cmd.ID() == commandID { // Check if it's the correct command
				c.groups = append(c.groups[:i], c.groups[i+1:]...) // Remove the group from the list
				check = cmd.GroupID()                              // Add the check as we need to remove the check from the client when the user isn't in the group anymore
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
	if i >= c.ChecksLength() || i < 0 {
		return false
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.checks = append(c.checks[:i], c.checks[i+1:]...)
	return true
}

// RemoveCheckByID will remove a specific check from the client based on the check ID
func (c *Client) RemoveCheckByID(id bson.ObjectId) (ok bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := len(c.checks) - 1; i >= 0; i-- {
		ch := c.checks[i]
		if ch.ID() == id {
			c.checks = append(c.checks[:i], c.checks[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveCheckByGroupID will remove a specific check from the client based on the group id
func (c *Client) RemoveCheckByGroupID(id bson.ObjectId) (ok bool) {
	ok = false
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := len(c.checks) - 1; i >= 0; i-- {
		ch := c.checks[i]
		if ch.Command().GroupID() == id {
			c.checks = append(c.checks[:i], c.checks[i+1:]...)
			ok = true
		}
	}
	return ok
}

// SendMessage will create a TCP connection to the client then send a message to it and return the reply or error
func (c Client) SendMessage(message string) (resp string, err error) {
	// Send the message to the client
	if _, err = fmt.Fprintln(c.conn, message); err != nil {
		return "", nil
	}

	// Read the reply
	r := bufio.NewReader(c.conn)
	msg, err := r.ReadString('\n')

	if err != nil {
		return "", err
	}
	return msg, nil
}

var re = regexp.MustCompile("-port=\"?([\\d,-]+)\"?") // special case check, when checking for port when pinging a server

// SendCheck will send a check to the client and then save the check to the database
func (c *Client) SendCheck(conn *nats.Conn, check *Check) string {
	// Start a check
	check.SetChecked(true)
	command := check.Command()

	if found, err := UpdateCheck(conn, check.ID()); err != nil || !found {
		if !found {
			ch := &models.Check{
				CommandID: check.Command().ID(),
				ClientID:  c.ID(),
				Response:  "",
				Checked:   check.Checked(),
				Error:     check.Error(),
				Finished:  check.Finished(),
			}
			ch.ID = check.ID()
			ch.CreatedAt = time.Now()
			ch.UpdatedAt = time.Now()

			if err = CreateCheck(conn, ch); err != nil {
				log.WithField("error", err).Error("error updating last check")
				return ""
			}
		} else {
			log.WithField("error", err).Error("error updating last check")
			return ""
		}
	}

	var resp string
	var err error

	// Special case when pinging as it should check if client is up and not send a command to the client
	if strings.HasPrefix(command.Command(), "ping") {
		// Check if port is provided in the command if not default port it 3333
		ports := ""
		p := re.FindStringSubmatch(command.Command())
		if len(p) >= 2 {
			ports = p[1]
		}

		if ports != "" {
			// Ping the server to check if its on or not
			r, err := c.Ping(ports)
			b, _ := json.Marshal(r)
			e := ""
			if err != nil {
				e = err.Error()
			}
			resp = fmt.Sprintf(`{"error":"%s","ports":%s}`, e, b)
		} else {
			resp, err = c.SendMessage("ping")
		}
	} else {
		// Send the command to the client and wait for a reply
		resp, err = c.SendMessage(command.Command())
	}

	// Check if there was an error in the connection or if the reply contains an error message
	if err != nil || strings.Contains(resp, `"error":""`) {
		// An error occured so put the check error to true and if it was a connection issue, set the response to the error message
		check.SetError(true)
		if err != nil {
			resp = err.Error()
		}
	} else {
		check.SetError(false)
	}
	resp = strings.TrimRight(resp, "\n")

	// Save the check to the database
	c.SaveCheck(conn, check, resp)
	return resp
}

// SaveCheck will save a check to the database
func (c *Client) SaveCheck(conn *nats.Conn, check *Check, resp string) {
	defer check.SetChecked(false)

	command := check.Command()

	ch := &models.Check{
		CommandID: command.ID(),
		ClientID:  c.ID(),
		Response:  resp,
		Error:     check.Error(),
		Finished:  true,
	}
	ch.ID = bson.NewObjectId()
	ch.CreatedAt = time.Now()
	ch.UpdatedAt = time.Now()

	if err := CreateCheck(conn, ch); err != nil {
		log.WithField("error", err).Error("error inserting new check")
		check.SetError(true)
		return
	}

	createdAt := ch.CreatedAt.Add(time.Duration(command.Interval()) * time.Second)
	check.SetTimestamp(createdAt)
	check.SetID(ch.ID)

	if !check.Error() && command.FailOnError() {
		c.ResetCheck(check.Group().Name())
	}

	for a := range check.IterAlerts() {
		a.Check(resp, conn)
	}
}

// ResetCheck will set error and checked to false on all checks with a specific name for the client
func (c *Client) ResetCheck(name string) {
	for ch := range c.IterChecks() {
		if ch.Group().Name() == name {
			ch.SetError(false)
			ch.SetChecked(false)
		}
	}
}

// StartCheck loop through all clients check and check if it's time to do any checks.
func (c *Client) StartCheck(conn *nats.Conn) {
	if c.Conn() == nil {
		return
	}

	// Start the loop
	for check := range c.IterChecks() {
		// If the check has already been started
		if check.Checked() {
			continue
		}

		// If the check contains any previous errors and the command has the condition to continue on error or not
		if check.Error() && !check.Command().FailOnError() {
			continue
		}

		// If its time for the next check or not
		now := time.Now()
		t := check.Timestamp()
		if z, _ := t.Zone(); z == "UTC" {
			t = t.In(now.Location()).Add(-(time.Hour * 4))
		}
		if !t.IsZero() && now.Before(t) {
			continue
		}

		// Send some debug message
		log.WithFields(log.Fields{
			"CommandID": check.Command().ID(),
			"ClientID":  c.ID(),
		}).Info("Starting a check for client")

		// Start a check
		go c.SendCheck(conn, check)
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
			conn, e := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.IP(), i), time.Duration(1)*time.Second)
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
