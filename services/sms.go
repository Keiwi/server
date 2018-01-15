package services

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Message    string      `json:"message"`
	Sender     string      `json:"sender"`
	Recipients []Recipient `json:"recipients"`
}

type Recipient struct {
	Msisdn string `json:"msisdn"`
}

type ServiceSMS struct {
	Recipients []Recipient
	GW         *GatewayAPI
}

func (ServiceSMS) Name() string {
	return "sms"
}

func (s ServiceSMS) Send(title string, msg string) {
	m := Message{
		Message:    msg,
		Sender:     title,
		Recipients: s.Recipients,
	}
	b, _ := json.Marshal(m)
	s.GW.SendSMS(string(b))
}

/** TODO Move this into its own package */
func newGatewayAPI(key string, timeout time.Duration) *GatewayAPI {
	return &GatewayAPI{
		APIKey: key,
		Base:   "https://GatewayAPI.com",
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

type GatewayAPI struct {
	Client *http.Client
	Base   string
	APIKey string
}

func (g GatewayAPI) SendSMS(message string) (*http.Response, error) {
	req, err := g.createRequest("/rest/mtsms?token="+g.APIKey, message)
	if err != nil {
		return nil, err
	}

	return g.Client.Do(req)
}

func (g GatewayAPI) createRequest(url string, data string) (*http.Request, error) {
	req, err := http.NewRequest("POST", g.Base+url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
