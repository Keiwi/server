package http

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/keiwi/server/virtual"
)

// CheckHandler is just the struct for a HTTP Handler
type CheckHandler struct{}

// Serve is the serve function for HTTP Handler when someone sends a web request
// This function will serve when someone wants to send a manual check
func (c CheckHandler) Serve(r *Handler) {
	// Try to check if there is a client with this ID in memory
	cl := r.GetServer().GetClients().GetClientByID(r.Request.ID)
	if cl == nil {
		r.OutputMessage(true, "Can't find a client in memory with this ID")
		return
	}

	// Check if a command ID was passed in the request
	if r.Request.Command == "" {
		// Check if the save param was passed in the request
		if r.Request.Save {
			// Loop through all checks with a specific command id and send them to the client
			// when running SendCheck on client, the check will also be saved to the database
			var resp []string
			for _, che := range cl.GetChecksByCommandID(r.Request.CommandID) {
				resp = append(resp, cl.SendCheck(r.GetServer().GetDatabase(), che))
			}

			// Parse all the response and output to the web response
			b, _ := json.Marshal(resp)
			r.OutputMessage(false, fmt.Sprintf("%s", b))
		} else {
			// Loop through all checks wiht a specific command id and send them to the client
			// when running SendCheck through CheckHandler it will send the data just like you
			// would when sending directly to the client struct, but it wont save it, it will
			// just keep the response in memory
			var resp []string
			for _, che := range cl.GetChecksByCommandID(r.Request.CommandID) {
				response, err := c.SendCheck(r, cl, che.GetCommand().GetCommand())
				if err != nil {
					r.OutputMessage(true, response)
				}
				resp = append(resp, response)
			}
			b, _ := json.Marshal(resp)
			r.OutputMessage(false, fmt.Sprintf("%s", b))
		}
	} else {
		// Command was supplied in web request instead of command ID, so just send the command directly to the client
		resp, err := c.SendCheck(r, cl, r.Request.Command)
		if err != nil {
			r.OutputMessage(true, resp)
		}
		r.OutputMessage(false, resp)
	}
}

// TODO: Maybe move this function directly into the client somehow to make things more understandable?

var re = regexp.MustCompile("-port=\"?([\\d,-]+)\"?") // special case check, when checking for port when pinging a GetServer().
// SendCheck will send a manual check to the client without saving to the database
func (c CheckHandler) SendCheck(r *Handler, cl *virtual.Client, command string) (string, error) {
	if strings.HasPrefix(command, "ping") {
		ports := "3333"
		p := re.FindStringSubmatch(r.Request.Command)
		if len(p) >= 2 {
			ports = p[1]
		}

		pingResponse, err := cl.Ping(ports)
		b, _ := json.Marshal(pingResponse)
		return string(b), err
	}

	resp, err := cl.SendMessage(command)
	if err != nil {
		return "Can't connect to client", err
	}
	return resp, nil
}
