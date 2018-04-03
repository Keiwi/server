package models

import (
	"sync"
	"time"

	"github.com/keiwi/utils/models"
	"gopkg.in/mgo.v2/bson"
)

// NewCheck - Creates a new virtual check
func NewCheck(ch *models.Check, command *Command) *Check {
	check := &Check{
		rw:      new(sync.RWMutex),
		command: command,
	}

	if ch == nil {
		return check
	}

	check.pastid = ch.ID
	check.checked = ch.Checked
	check.err = ch.Error
	check.finished = ch.Finished
	check.nexttimestamp = ch.CreatedAt
	return check
}

// Check - Virtual check struct
type Check struct {
	rw            *sync.RWMutex
	command       *Command
	group         *Group
	alerts        []*Alert
	pastid        bson.ObjectId
	nexttimestamp time.Time
	checked       bool
	err           bool
	finished      bool
}

// Command returns the command
func (c *Check) Command() (cmd *Command) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.command
}

// Group returns the group that this check belongs to
func (c *Check) Group() (g *Group) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.group
}

// Alerts returns all of the alerts associated with this check
func (c *Check) Alerts() []*Alert {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.alerts
}

// ID returns the checks ID that this check belongs to
func (c *Check) ID() (id bson.ObjectId) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.pastid
}

// Timestamp returns the checks timestamp
func (c *Check) Timestamp() (timestamp time.Time) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.nexttimestamp
}

// Checked returns whether or not the check has been checked.
func (c *Check) Checked() (checked bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.checked
}

// Error returns whether or not the check got an error
func (c *Check) Error() (err bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.err
}

// Finished returns whether or not the check is finished
func (c *Check) Finished() (finished bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.finished
}

// SetGroup modifies the group that this check belongs to
func (c *Check) SetGroup(g *Group) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.group = g
}

// SetID modifies the check id
func (c *Check) SetID(id bson.ObjectId) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.pastid = id
}

// SetTimestamp modifies the checks timestamp
func (c *Check) SetTimestamp(t time.Time) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.nexttimestamp = t
}

// SetChecked modifies whether the check has been checked or not
func (c *Check) SetChecked(checked bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.checked = checked
}

// SetError modifies whether the check got an error or not
func (c *Check) SetError(err bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.err = err
}

// SetFinished modifies whether the check is finished or not
func (c *Check) SetFinished(finished bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.finished = finished
}

// AddAlert will add a new alert to the check
func (c *Check) AddAlert(alert *Alert) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.alerts = append(c.alerts, alert)
}

// RemoveAlertByID will remove an alert from the check
func (c *Check) RemoveAlertByID(id bson.ObjectId) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := len(c.alerts) - 1; i >= 0; i-- {
		a := c.alerts[i]
		if a.ID() == id {
			c.alerts = append(c.alerts[:i], c.alerts[i+1:]...)
			return
		}
	}
}

// AlertsLength will return the amount of alerts associated with this check
func (c Check) AlertsLength() int {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.alerts)
}

// IterServices - Will return a channel and loop through all the alerts in a safe way and pass it to the channel
func (c Check) IterAlerts() <-chan *Alert {
	ch := make(chan *Alert, c.AlertsLength())
	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()
		for _, alert := range c.alerts {
			ch <- alert
		}
		close(ch)
	}()
	return ch
}
