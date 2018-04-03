package models

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/keiwi/server/providers"
	"github.com/keiwi/server/services"
	"github.com/keiwi/utils/models"
	"github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2/bson"
)

// NewAlert - will create a new virtual alert based on AlertOptions
func NewAlert(ao models.AlertOption) *Alert {
	alert := &Alert{
		rw:            new(sync.RWMutex),
		id:            ao.ID,
		clientid:      ao.ClientID,
		delay:         ao.Delay,
		previousalert: time.Time{},
	}

	switch ao.Alert {
	case "cpu":
		avg, err := strconv.ParseFloat(ao.Value, 64)
		if err != nil {
			return nil
		}
		alert.alert = providers.NewAlertProviderCPU(ao.Count, avg)
	}

	for _, s := range strings.Split(ao.Service, ",") {
		switch s {
		case "sms":
			alert.services = append(alert.services, services.NewServiceSMS())
		case "email":
			// TODO: Implement emails
		}
	}
	return alert
}

// Alert - is the virtual alert struct
type Alert struct {
	rw            *sync.RWMutex
	id            bson.ObjectId
	clientid      bson.ObjectId
	delay         int
	alert         providers.AlertProvider
	previousalert time.Time
	services      []services.Service
}

// ID - Will return the alert ID
func (a Alert) ID() bson.ObjectId {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.id
}

// ClientID - Will return the client id on the alert
func (a Alert) ClientID() bson.ObjectId {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.clientid
}

// Delay - Will return the delay between alerts
func (a Alert) Delay() int {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.delay
}

// Alert - Will return the AlertProvider
func (a Alert) Alert() providers.AlertProvider {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.alert
}

// PreviousAlert - Will return when the previous alert was made
func (a Alert) PreviousAlert() time.Time {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.previousalert
}

// Services - Will return all of the services associated with the alert
func (a Alert) Services() []services.Service {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.services
}

// SetID - Will modify the virtual ID
func (a *Alert) SetID(id bson.ObjectId) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.id = id
}

// SetClientID Will modify the virtual Client ID
func (a *Alert) SetClientID(clientID bson.ObjectId) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.clientid = clientID
}

// SetDelay - Will modify the virtual delay
func (a *Alert) SetDelay(delay int) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.delay = delay
}

// SetAlert - Will modify the AlertProvider associated with the alert
func (a *Alert) SetAlert(alert providers.AlertProvider) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.alert = alert
}

// SetPreviousAlert - Will modify when the previous alert was made
func (a *Alert) SetPreviousAlert(previous time.Time) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.previousalert = previous
}

// SetServices - Will modify all of the services associated with the alert
func (a *Alert) SetServices(s []services.Service) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.services = s
}

// Update - will update the virtual alert data based on AlertOptions
func (a *Alert) Update(alert models.AlertOption) bool {
	if alert.Delay != a.Delay() {
		a.SetDelay(alert.Delay)
	}

	switch alert.Alert {
	case "cpu":
		avg, err := strconv.ParseFloat(alert.Value, 64)
		if err != nil {
			return false
		}
		if strings.ToLower(a.Alert().Name()) != alert.Alert {
			a.SetAlert(providers.NewAlertProviderCPU(alert.Count, avg))
		} else {
			if a.Alert().(*providers.CPU).Total != alert.Count {
				a.Alert().(*providers.CPU).Total = alert.Count
			}
			if a.Alert().(*providers.CPU).Avg != avg {
				a.Alert().(*providers.CPU).Avg = avg
			}
		}
	}

	a.rw.Lock()
	defer a.rw.Unlock()
	a.services = []services.Service{}
	for _, s := range strings.Split(alert.Service, ",") {
		switch s {
		case "sms":
			a.services = append(a.services, services.NewServiceSMS())
		case "email":
			// TODO: Implement this
		}
	}
	return true
}

// Check - Will check if an alert should be made or not
func (a *Alert) Check(resp string, conn *nats.Conn) {
	a.rw.RLock()
	al := a.alert
	a.rw.RUnlock()
	if al.Check(resp) {
		if time.Now().After(a.PreviousAlert()) {
			for s := range a.IterServices() {
				s.Send(al.Name(), al.Message())
			}
		}

		alert := models.Alert{
			AlertID:  a.ID(),
			ClientID: a.ClientID(),
			Value:    al.Value(),
		}
		alert.UpdatedAt = time.Now()
		alert.CreatedAt = time.Now()

		data, err := bson.MarshalJSON(alert)
		if err != nil {
			a.SetPreviousAlert(time.Now().Add(time.Duration(a.delay) * time.Second))
			return
		}

		err = conn.Publish("alerts.create.send", data)
		if err != nil {
			a.SetPreviousAlert(time.Now().Add(time.Duration(a.delay) * time.Second))
			return
		}

		a.SetPreviousAlert(alert.CreatedAt)
	}
}

// ServicesLength - Will return the amount of services associated with the alert
func (a Alert) ServicesLength() int {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return len(a.services)
}

// IterServices - Will return a channel and loop through all the services in a safe way and pass it to the channel
func (a Alert) IterServices() <-chan services.Service {
	ch := make(chan services.Service, a.ServicesLength())
	go func() {
		a.rw.RLock()
		defer a.rw.RUnlock()
		for _, service := range a.services {
			ch <- service
		}
		close(ch)
	}()
	return ch
}
