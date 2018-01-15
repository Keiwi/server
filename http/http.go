package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/keiwi/server/database"
	"github.com/keiwi/server/virtual"
	"github.com/keiwi/utils"
)

// APIRequest is the struct that will be used when someone sends a request to the aPI
type APIRequest struct {
	Type      string `json:"type"`
	Command   string `json:"command"`
	GroupName string `json:"group_name"`
	ID        uint   `json:"id"`
	CommandID uint   `json:"command_id"`
	GroupID   uint   `json:"group_id"`
	Save      bool   `json:"save"`
}

// Handler is the struct for handling all the API request
type Handler struct {
	Request APIRequest
	s       Server
	w       http.ResponseWriter
	r       *http.Request
}

// GetServer will return the HTTP server struct
func (h Handler) GetServer() Server {
	return h.s
}

// Output is the function for outputting data to the web request
func (h Handler) Output(a APIResponse) {
	out, err := json.Marshal(a)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		http.Error(h.w, "Something went wrong internal, contact IT support", http.StatusInternalServerError)
		return
	}
	h.w.Header().Set("Content-Type", "application/json")
	h.w.Write(out)
}

// OutputMessage is a wrapper for Output for outputting data
func (h Handler) OutputMessage(err bool, message string) {
	h.Output(APIResponse{Error: err, Message: message})
}

// APIHandler is an interface for the API
type APIHandler interface {
	Insert(*Handler)
	Update(*Handler)
	Delete(*Handler)
}

// APIResponse is the struct when outputting data
type APIResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// Server is the struct for API server
type Server struct {
	Handlers map[string]APIHandler
	Check    CheckHandler
	Clients  *virtual.Clients
	Database *database.Database
}

// Start will handle the startup of the API, add all the handlers, start the HTTP Server
func (h *Server) Start(adress string) {
	h.Handlers["command"] = CommandHandler{}
	h.Handlers["client"] = ClientHandler{}
	h.Handlers["group"] = GroupHandler{}
	h.Handlers["alert"] = GroupHandler{}
	h.Check = CheckHandler{}

	http.Handle("/", h)
	err := http.ListenAndServe(adress, nil)
	if err != nil {
		utils.Log.WithField("error", err).Error("Can't start web API")
	}
}

// ServeHTTP will handle the main HTTP request and redirect to other functions depending on the request
func (h Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiHandler := &Handler{
		s: h, // HTTP Server
		w: w, // ResponseWriter
		r: r, // HTTP request
	}

	decoder := json.NewDecoder(r.Body)
	req := APIRequest{}
	err := decoder.Decode(&req)
	if err != nil {
		utils.Log.WithField("error", err).Error("Internal error")
		apiHandler.Output(APIResponse{Error: true, Message: "Internal error when parsing reqest body"})
		return
	}
	defer r.Body.Close()
	apiHandler.Request = req

	if strings.HasPrefix(r.URL.Path, "/api/check") {
		h.Check.Serve(apiHandler)
	}
	if strings.HasPrefix(r.URL.Path, "/api") {
		path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(path) < 2 {
			apiHandler.Output(APIResponse{Error: true, Message: "Invalid Path"})
			return
		}
		handler, ok := h.Handlers[path[1]]
		if !ok {
			apiHandler.Output(APIResponse{Error: true, Message: "Invalid Path"})
			return
		}

		switch strings.ToLower(req.Type) {
		case "insert":
			handler.Insert(apiHandler)
			break
		case "update":
			handler.Update(apiHandler)
			break
		case "delete":
			handler.Delete(apiHandler)
			break
		default:
			apiHandler.Output(APIResponse{Error: true, Message: "Invalid Request type"})
			return
		}
	}
}

// GetClients will return all of the clients
func (h Server) GetClients() *virtual.Clients {
	return h.Clients
}

// GetDatabase will return the database instance
func (h Server) GetDatabase() *database.Database {
	return h.Database
}
