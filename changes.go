package server

import (
	"errors"
	"time"

	"github.com/keiwi/server/models"
	"github.com/keiwi/utils"

	db "github.com/keiwi/utils/models"

	"github.com/nats-io/go-nats"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2/bson"
)

func FindAlerts(conn *nats.Conn, filter utils.Filter) ([]db.Alert, error) {
	requestData := utils.FindOptions{
		Filter: filter,
		Sort:   utils.Sort{"created_at"},
	}
	data, err := bson.MarshalJSON(requestData)
	if err != nil {
		return nil, err
	}
	msg, err := conn.Request("checks.retrieve.find", data, time.Duration(viper.GetInt("nats_delay"))*time.Second)
	if err != nil {
		return nil, err
	}

	var alerts []db.Alert
	err = bson.UnmarshalJSON(msg.Data, &alerts)
	if err != nil {
		return nil, err
	}

	return alerts, nil
}

func FindWithAlertAndClient(conn *nats.Conn, alert, client bson.ObjectId) (*db.Alert, error) {
	alerts, err := FindAlerts(conn, utils.Filter{"alert_id": alert, "client_id": client})
	if err != nil {
		return nil, err
	}

	if len(alerts) <= 0 {
		return nil, errors.New("could not find any alerts")
	}
	return &alerts[0], nil
}

func handleDatabaseChanges() {
	natsConn.Subscribe("alert_options.update.after", func(m *nats.Msg) {
		var alert db.AlertOption
		err := bson.UnmarshalJSON(m.Data, &alert)
		if err != nil {
			utils.Log.WithError(err).Errorf("error decoding event (%s)", "alerts.update")
			return
		}

		for cl := range manager.IterClients() {
			if cl.ID() == alert.ClientID {
				for ch := range cl.IterChecks() {
					if ch.Command().ID() == alert.CommandID {
						for a := range ch.IterAlerts() {
							if a.ID() == alert.ID {
								a.Update(alert)
							}
						}
						return
					}
				}
			}
		}
	})

	natsConn.Subscribe("alert_options.delete.after", func(m *nats.Msg) {
		var alerts []db.AlertOption
		err := bson.UnmarshalJSON(m.Data, &alerts)
		if err != nil {
			utils.Log.WithError(err).Errorf("error decoding event (%s)", "alerts.update")
			return
		}

		for cl := range manager.IterClients() {
			for ch := range cl.IterChecks() {
				for alert := range ch.IterAlerts() {
					for _, a := range alerts {
						if alert.ID() == a.ID {
							ch.RemoveAlertByID(a.ID)
							break
						}
					}
				}
			}
		}
	})

	natsConn.Subscribe("alert_options.create.after", func(m *nats.Msg) {
		var alert db.AlertOption
		err := bson.UnmarshalJSON(m.Data, &alert)
		if err != nil {
			utils.Log.WithError(err).Errorf("error decoding event (%s)", "alerts.update")
			return
		}

		for cl := range manager.IterClients() {
			for ch := range cl.IterChecks() {
				for al := range ch.IterAlerts() {
					if al.ID() == alert.ID {
						return
					}
				}
			}
		}

		for cl := range manager.IterClients() {
			if cl.ID() == alert.ClientID {
				for ch := range cl.IterChecks() {
					if ch.Command().ID() == alert.CommandID {
						a := models.NewAlert(alert)

						al, err := FindWithAlertAndClient(natsConn, a.ID(), cl.ID())
						if err == nil {
							a.SetPreviousAlert(al.CreatedAt)
						}
						return
					}
				}
			}
		}
	})

	natsConn.Subscribe("checks.delete.after", func(m *nats.Msg) {
		var checks []db.Check
		err := bson.UnmarshalJSON(m.Data, &checks)
		if err != nil {
			utils.Log.WithError(err).Errorf("error decoding event (%s)", "checks.update")
			return
		}

		for cl := range manager.IterClients() {
			for _, ch := range checks {
				cl.RemoveCheckByID(ch.ID)
			}
		}
	})

	natsConn.Subscribe("clients.update.after", func(m *nats.Msg) {
		var client db.Client
		err := bson.UnmarshalJSON(m.Data, &client)
		if err != nil {
			utils.Log.WithError(err).Errorf("error decoding event (%s)", "clients.update")
			return
		}

		// Find the client in memory
		cl := manager.ClientByID(client.ID)

		// If client does not exists in memory then it was newly created, if so add to memory
		if cl == nil {
			c := models.NewClient(&client)
			manager.AddClient(c)
		} else {
			cl.SetIP(client.IP)
		}
	})
	natsConn.Subscribe("clients.delete.after", func(m *nats.Msg) {
		var clients []db.Client
		err := bson.UnmarshalJSON(m.Data, &clients)
		if err != nil {
			utils.Log.WithError(err).Errorf("error decoding event (%s)", "clients.update")
			return
		}

		for _, cl := range clients {
			manager.RemoveClientByID(cl.ID)
		}
	})
}
