package services

import "time"

type Service interface {
	Send(string, string)
	Name() string
}

func NewServiceSMS() Service {
	return &ServiceSMS{
		Recipients: []Recipient{
			Recipient{Msisdn: "46724206544"},
		},
		GW: newGatewayAPI("_qVpWmD_Q-OyvBy47KdFO1nSI9s7qxRUwimw7-Anjbw_HEMiRkMer4RudKAyhE9H", time.Second*10),
	}
}
