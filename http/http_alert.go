package http

import (
	"github.com/keiwi/server/virtual"
	"github.com/keiwi/utils"
)

// AlertHandler is just the struct for a HTTP Handler
type AlertHandler struct{}

// Insert is ran when someone tries to create a new alert through the API
func (AlertHandler) Insert(r *Handler) {
	alert, err := r.GetServer().GetDatabase().AlertOptions.Find(r.Request.ID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.Output(APIResponse{Error: true, Message: "Internal error"})
		return
	}

	for cl := range r.GetServer().GetClients().IterClients() {
		for ch := range cl.IterChecks() {
			for al := range ch.IterAlerts() {
				if al.GetID() == r.Request.ID {
					r.Output(APIResponse{Error: true, Message: "There is already an alert with this ID in cache."})
					return
				}
			}
		}
	}

	for cl := range r.GetServer().GetClients().IterClients() {
		if cl.GetID() == alert.ClientID {
			for ch := range cl.IterChecks() {
				if ch.GetCommand().GetID() == alert.CommandID {
					a := virtual.NewAlert(alert)

					al, err := r.GetServer().GetDatabase().Alerts.FindWithAlertAndClient(a.GetID(), cl.GetID())
					if err == nil {
						a.SetPreviousAlert(al.CreatedAt)
					}
					r.Output(APIResponse{Error: false, Message: "Added the alert to the client"})
					return
				}
			}
		}
	}
	r.Output(APIResponse{Error: true, Message: "Can't find the correct client or command to add the alert to"})
}

// Update is ran when someone tries to update some data on an alert
func (AlertHandler) Update(r *Handler) {
	alert, err := r.GetServer().GetDatabase().AlertOptions.Find(r.Request.ID)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		r.Output(APIResponse{Error: true, Message: "Internal error"})
		return
	}

	for cl := range r.GetServer().GetClients().IterClients() {
		if cl.GetID() == alert.ClientID {
			for ch := range cl.IterChecks() {
				if ch.GetCommand().GetID() == alert.CommandID {
					for a := range ch.IterAlerts() {
						if a.GetID() == alert.ID {
							a.Update(alert)
						}
					}
					r.Output(APIResponse{Error: false, Message: "Updated the alert for the client"})
					return
				}
			}
		}
	}
	r.Output(APIResponse{Error: true, Message: "Can't find the correct client or command to update the alert"})
}

// Delete is ran when someone tries to delete an alert
func (AlertHandler) Delete(r *Handler) {
	for cl := range r.GetServer().GetClients().IterClients() {
		for ch := range cl.IterChecks() {
			for alert := range ch.IterAlerts() {
				if alert.GetID() == r.Request.ID {
					ch.RemoveAlertByID(r.Request.ID)
					r.Output(APIResponse{Error: false, Message: "Deleted the alert from the client"})
					return
				}
			}
		}
	}
	r.Output(APIResponse{Error: true, Message: "Can't find the correct client or command to delete the alert from"})
}
