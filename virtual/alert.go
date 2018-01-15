package virtual

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/keiwi/server/database"
	"github.com/keiwi/server/providers"
	"github.com/keiwi/server/services"
)

// NewAlert - will create a new virtual alert based on AlertOptions
func NewAlert(ao database.AlertOptions) *Alert {
	alert := &Alert{
		rw:            new(sync.RWMutex),
		ID:            ao.ID,
		ClientID:      ao.ClientID,
		Delay:         ao.Delay,
		PreviousAlert: time.Time{},
	}

	switch ao.Alert {
	case "cpu":
		avg, err := strconv.ParseFloat(ao.Value, 64)
		if err != nil {
			return nil
		}
		alert.Alert = providers.NewAlertProviderCPU(ao.Count, avg)
	}

	for _, s := range strings.Split(ao.Service, ",") {
		switch s {
		case "sms":
			alert.Services = append(alert.Services, services.NewServiceSMS())
		case "email":
			// TODO: Implement emails
		}
	}
	return alert
}

// Alert - is the virtual alert struct
type Alert struct {
	rw            *sync.RWMutex
	ID            uint
	ClientID      uint
	Delay         uint
	Alert         providers.AlertProvider
	PreviousAlert time.Time
	Services      []services.Service
}

// GetID - Will return the alert ID
func (a Alert) GetID() uint {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.ID
}

// GetClientID - Will return the client id on the alert
func (a Alert) GetClientID() uint {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.ClientID
}

// GetDelay - Will return the delay between alerts
func (a Alert) GetDelay() uint {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.Delay
}

// GetAlert - Will return the AlertProvider
func (a Alert) GetAlert() providers.AlertProvider {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.Alert
}

// GetPreviousAlert - Will return when the previous alert was made
func (a Alert) GetPreviousAlert() time.Time {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.PreviousAlert
}

// GetServices - Will return all of the services associated with the alert
func (a Alert) GetServices() []services.Service {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.Services
}

// SetID - Will modify the virtual ID
func (a *Alert) SetID(id uint) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.ID = id
}

// SetClientID Will modify the virtual Client ID
func (a *Alert) SetClientID(clientID uint) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.ClientID = clientID
}

// SetDelay - Will modify the virtual delay
func (a *Alert) SetDelay(delay uint) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.Delay = delay
}

// SetAlert - Will modify the AlertProvider associated with the alert
func (a *Alert) SetAlert(alert providers.AlertProvider) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.Alert = alert
}

// SetPreviousAlert - Will modify when the previous alert was made
func (a *Alert) SetPreviousAlert(previous time.Time) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.PreviousAlert = previous
}

// SetServices - Will modify all of the services associated with the alert
func (a *Alert) SetServices(s []services.Service) {
	a.rw.Lock()
	defer a.rw.Unlock()
	a.Services = s
}

// Update - will update the virtual alert data based on AlertOptions
func (a *Alert) Update(alert database.AlertOptions) bool {
	if alert.Delay != a.GetDelay() {
		a.SetDelay(alert.Delay)
	}

	switch alert.Alert {
	case "cpu":
		avg, err := strconv.ParseFloat(alert.Value, 64)
		if err != nil {
			return false
		}
		if strings.ToLower(a.GetAlert().Name()) != alert.Alert {
			a.SetAlert(providers.NewAlertProviderCPU(alert.Count, avg))
		} else {
			if a.GetAlert().(*providers.AlertProviderCPU).Total != alert.Count {
				a.GetAlert().(*providers.AlertProviderCPU).Total = alert.Count
			}
			if a.GetAlert().(*providers.AlertProviderCPU).Avg != avg {
				a.GetAlert().(*providers.AlertProviderCPU).Avg = avg
			}
		}
	}

	a.rw.Lock()
	defer a.rw.Unlock()
	a.Services = []services.Service{}
	for _, s := range strings.Split(alert.Service, ",") {
		switch s {
		case "sms":
			a.Services = append(a.Services, services.NewServiceSMS())
		case "email":
			// TODO: Implement this
		}
	}
	return true
}

// Check - Will check if an alert should be made or not
func (a *Alert) Check(resp string, db *database.Database) {
	a.rw.RLock()
	al := a.Alert
	a.rw.RUnlock()
	if al.Check(resp) {
		if time.Now().After(a.GetPreviousAlert()) {
			for s := range a.IterServices() {
				s.Send(a.Alert.Name(), a.Alert.Message())
			}
		}

		alert := database.Alert{
			AlertID:  a.GetID(),
			ClientID: a.GetClientID(),
			Value:    al.Value(),
		}

		err := db.Alerts.Save(&alert)
		if err != nil {
			a.rw.Lock()
			a.PreviousAlert = time.Now().Add(time.Duration(a.Delay) * time.Second)
			a.rw.Unlock()
		} else {
			a.PreviousAlert = alert.CreatedAt
		}
	}
}

// CountServices - Will return the amount of services associated with the alert
func (a Alert) CountServices() int {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return len(a.Services)
}

// IterServices - Will return a channel and loop through all the services in a safe way and pass it to the channel
func (a Alert) IterServices() <-chan services.Service {
	ch := make(chan services.Service, a.CountServices())
	go func() {
		a.rw.RLock()
		defer a.rw.RUnlock()
		for _, service := range a.Services {
			ch <- service
		}
		close(ch)
	}()
	return ch
}
