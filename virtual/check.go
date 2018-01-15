package virtual

import (
	"sync"
	"time"

	"github.com/keiwi/server/database"
)

// NewCheck - Creates a new virtual check
func NewCheck(ch *database.Check, command *Command) *Check {
	check := &Check{
		rw:            new(sync.RWMutex),
		Command:       command,
		PastID:        ch.ID,
		Checked:       ch.Checked,
		Err:           ch.Error,
		Finished:      ch.Finished,
		NextTimestamp: ch.CreatedAt,
	}
	return check
}

// Check - Virtual check struct
type Check struct {
	rw            *sync.RWMutex
	Command       *Command
	Group         *Group
	Alerts        []*Alert
	PastID        uint
	NextTimestamp time.Time
	Checked       bool
	Err           bool
	Finished      bool
}

// GetCommand returns the command
func (c *Check) GetCommand() (cmd *Command) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	cmd = c.Command
	return
}

// GetGroup returns the group that this check belongs to
func (c *Check) GetGroup() (g *Group) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	g = c.Group
	return
}

// GetAlerts returns all of the alerts associated with this check
func (c *Check) GetAlerts() []*Alert {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.Alerts
}

// GetID returns the checks ID that this check belongs to
func (c *Check) GetID() (id uint) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	id = c.PastID
	return
}

// GetTimestamp returns the checks timestamp
func (c *Check) GetTimestamp() (timestamp time.Time) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	timestamp = c.NextTimestamp
	return
}

// GetChecked returns whether or not the check has been checked.
func (c *Check) GetChecked() (checked bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	checked = c.Checked
	return
}

// GetError returns whether or not the check got an error
func (c *Check) GetError() (err bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	err = c.Err
	return
}

// GetFinished returns whether or not the check is finished
func (c *Check) GetFinished() (finished bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	finished = c.Finished
	return
}

// SetGroup modifies the group that this check belongs to
func (c *Check) SetGroup(g *Group) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Group = g
}

// SetID modifies the check id
func (c *Check) SetID(id uint) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.PastID = id
}

// SetTimestamp modifies the checks timestamp
func (c *Check) SetTimestamp(t time.Time) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.NextTimestamp = t
}

// SetChecked modifies whether the check has been checked or not
func (c *Check) SetChecked(checked bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Checked = checked
}

// SetError modifies whether the check got an error or not
func (c *Check) SetError(err bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Err = err
}

// SetFinished modifies whether the check is finished or not
func (c *Check) SetFinished(finished bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Finished = finished
}

// AddAlert will add a new alert to the check
func (c *Check) AddAlert(alert *Alert) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.Alerts = append(c.Alerts, alert)
}

// RemoveAlertByID will remove an alert from the check
func (c *Check) RemoveAlertByID(id uint) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := len(c.Alerts) - 1; i >= 0; i-- {
		a := c.Alerts[i]
		if a.GetID() == id {
			c.Alerts = append(c.Alerts[:i], c.Alerts[i+1:]...)
			return
		}
	}
}

// CountAlerts will return the amount of alerts associated with this check
func (c Check) CountAlerts() int {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.Alerts)
}

// IterServices - Will return a channel and loop through all the alerts in a safe way and pass it to the channel
func (c Check) IterAlerts() <-chan *Alert {
	ch := make(chan *Alert, c.CountAlerts())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, alert := range c.Alerts {
			ch <- alert
		}
		close(ch)
	}()
	return ch
}
