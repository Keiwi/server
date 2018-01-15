package server

import (
	"os"
	"strings"

	"github.com/keiwi/server/http"
	"github.com/keiwi/utils"
	"github.com/spf13/viper"

	"time"

	"github.com/keiwi/server/database"
	"github.com/keiwi/server/virtual"
)

// Server is the struct for the actual server
type Server struct {
	clients *virtual.Clients

	database *database.Database
}

// Start is the function for starting the server and initialize everything
func (s *Server) Start() {
	ReadConfig()

	utils.Log.Info("Starting keiwi Monitor server")

	// Initialize the database connection
	utils.Log.Info("Connecting to SQL database")
	db, err := database.NewMysqlDB(
		viper.GetString("mysql_username"),
		viper.GetString("mysql_password"),
		viper.GetString("mysql_host"),
		viper.GetString("mysql_port"),
		viper.GetString("mysql_database"))
	if err != nil {
		utils.Log.WithField("error", err).Fatal("can't connect to database")
	}
	s.database = db

	// Get all the clients from database and create virtual clients of them
	utils.Log.Info("Creating all clients")
	clients, err := virtual.NewClients(s.database)
	if err != nil {
		s.database.Close()
		utils.Log.WithField("error", err).Fatal("Something went wrong when creating all the clients")
		return
	}
	s.clients = clients
	utils.Log.Info("Finished creating all clients")

	utils.Log.Info("Starting web API")
	API := http.Server{
		Clients:  s.GetClients(),
		Database: s.GetDatabase(),
		Handlers: map[string]http.APIHandler{},
	}
	go API.Start(viper.GetString("api_adress"))
	utils.Log.Info("Web API started")

	// Start the timer loop
	utils.Log.Info("Starting loop")
	for {
		go s.Loop()
		time.Sleep(time.Second * time.Duration(viper.GetInt("interval")))
	}
}

// Loop is the function for the timer, it will continiously loop all the clients and check if its time for a check
func (s *Server) Loop() {
	for cl := range s.clients.IterClients() {
		go cl.Check(s.GetDatabase())
	}
}

// GetDatabase will return the database instance
func (s Server) GetDatabase() *database.Database {
	return s.database
}

// GetClients will return all of the clients
func (s Server) GetClients() *virtual.Clients {
	return s.clients
}

// ReadConfig will try to find the config and read, if config file
// does not exists it will create one with default options
func ReadConfig() {
	configType := os.Getenv("KeiwiServerConfigType")
	if configType == "" {
		configType = "json"
	}
	viper.SetConfigType(configType)

	viper.SetConfigFile("config." + configType)
	viper.AddConfigPath(".")

	viper.SetDefault("log_dir", "./logs")
	viper.SetDefault("log_syntax", "%date%_server.log")
	viper.SetDefault("log_level", "info")

	viper.SetDefault("mysql_username", "root")
	viper.SetDefault("mysql_password", "")
	viper.SetDefault("mysql_host", "127.0.0.1")
	viper.SetDefault("mysql_port", "3306")
	viper.SetDefault("mysql_database", "")

	viper.SetDefault("api_adress", ":8080")
	viper.SetDefault("interval", 1)

	if err := viper.ReadInConfig(); err != nil {
		utils.Log.Debug("Config file not found, saving default")
		if err = viper.WriteConfigAs("config." + configType); err != nil {
			utils.Log.WithField("error", err.Error()).Fatal("Can't save default config")
		}
	}

	level := strings.ToLower(viper.GetString("log_level"))
	utils.Log = utils.NewLogger(utils.NameToLevel[level], &utils.LoggerConfig{
		Dirname: viper.GetString("log_dir"),
		Logname: viper.GetString("log_syntax"),
	})
}
