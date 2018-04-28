package server

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/keiwi/server/models"
	"github.com/keiwi/utils/log"
	"github.com/keiwi/utils/log/handlers/cli"
	"github.com/keiwi/utils/log/handlers/file"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	configType string
	manager    *models.Manager
	natsConn   *nats.Conn
	tcpConn    net.Listener
	kill       bool
)

func Start() {
	ReadConfig()

	log.Info("Starting keiwi Monitor Client")

	log.Info("Connecting to nats")
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.WithError(err).Error("error connecting to nats")
		return
	}
	natsConn = c
	log.Info("Nats connection successful")

	// Get all the clients from database and create virtual clients of them
	log.Info("Creating the manager")
	man, err := models.NewManager(natsConn)
	if err != nil {
		Close()
		log.WithError(err).Fatal("Something went wrong when creating the manager")
		return
	}
	manager = man
	log.Info("Finished creating the manager")

	log.Info("Starting to listen for database changes")
	handleDatabaseChanges()
	log.Info("Listening for database changes")

	log.Info("Starting loop")
	go func() {
		for {
			if kill {
				return
			}
			go Loop(natsConn)
			time.Sleep(time.Second * time.Duration(viper.GetInt("interval")))
		}
	}()
	log.Info("Loop started")

	log.Info("Configuring certificates")
	cer, err := tls.LoadX509KeyPair("D:/ssh/server.crt", "D:/ssh/server.key")
	if err != nil {
		Close()
		log.WithError(err).Error("error loading certificate")
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Info("Starting TCP server")
	s, err := tls.Listen("tcp", viper.GetString("server_ip"), config)
	if err != nil {
		Close()
		log.WithError(err).Error("error starting TLS server")
		return
	}
	log.WithField("ip", s.Addr().String()).Info("TCP server started")

	log.Info("Waiting for clients")
	for {
		conn, err := s.Accept()
		if err != nil {
			log.WithError(err).Error("error accepting TLS request")
			continue
		}
		log.WithField("connection", conn.RemoteAddr().String()).Info("TCP client trying to connect")

		go handleConnection(conn)
	}
}

// Loop is the function for the timer, it will continiously loop all the clients and check if its time for a check
func Loop(conn *nats.Conn) {
	for cl := range manager.IterClients() {
		go cl.StartCheck(conn)
	}
}

func Close() {
	kill = true
	if natsConn != nil {
		natsConn.Close()
	}
	if tcpConn != nil {
		tcpConn.Close()
	}
}

// ReadConfig will try to find the config and read, if config file
// does not exists it will create one with default options
func ReadConfig() {
	configType = os.Getenv("KeiwiConfigType")
	if configType == "" {
		configType = "json"
	}
	viper.SetConfigType(configType)

	viper.SetConfigFile("config." + configType)
	viper.AddConfigPath(".")

	viper.SetDefault("log_dir", "./logs")
	viper.SetDefault("log_syntax", "%date%_server.log")
	viper.SetDefault("log_level", "info")

	viper.SetDefault("server_ip", "127.0.0.1:4444")
	viper.SetDefault("password", GenerateRandomString(32))
	viper.SetDefault("interval", 600)
	viper.SetDefault("nats_delay", 10)

	if err := viper.ReadInConfig(); err != nil {
		log.Debug("Config file not found, saving default")
		if err = viper.WriteConfigAs("config." + configType); err != nil {
			log.WithField("error", err.Error()).Fatal("Can't save default config")
		}
	}

	fileConfig := file.Config{Folder: viper.GetString("log_dir"), Filename: viper.GetString("log_syntax")}

	level := strings.ToLower(viper.GetString("log_level"))
	log.Log = log.NewLogger(log.GetLevelFromString(level), []log.Reporter{
		cli.NewCli(),
		file.NewFile(&fileConfig),
	})
}

func handleConnection(conn net.Conn) {
	accepted, err := Handshake(conn)
	if err != nil {
		conn.Close()
		log.WithError(err).Error("handshake failed")
		return
	}

	if !accepted {
		conn.Close()
		log.WithField("ip", conn.RemoteAddr().String()).Info("tcp handshake declined")
		return
	}
	log.WithField("ip", conn.RemoteAddr().String()).Info("tcp handshake accepted")

	ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	for cl := range manager.IterClients() {
		if cl.IP() == ip {
			cl.SetConn(conn)
		}
	}
}

func Handshake(conn net.Conn) (bool, error) {
	r := bufio.NewReader(conn)
	password, err := r.ReadString('\n')
	if err != nil {
		return false, errors.Wrap(err, "connection disconnected")
	}
	password = strings.TrimSpace(password)

	log.WithField("server_password", viper.GetString("password")).WithField("client_password", password).Info("Checking password")

	if password != viper.GetString("password") {
		_, err := fmt.Fprintln(conn, "declined")
		return false, err
	}

	_, err = fmt.Fprintln(conn, "accepted")
	if err != nil {
		return false, err
	}
	return true, nil
}

// GenerateRandomKey is a wrapper over securecookie.GenerateRandomKey to generate a string
// rather then a byte slice
func GenerateRandomString(len int) string {
	// Generate random key from securecookie
	data := GenerateRandomKey(len)
	if data == nil {
		return ""
	}

	// Encode the random key
	return base64.StdEncoding.EncodeToString(data)
}

// GenerateRandomKey creates a random key with the given length in bytes.
// On failure, returns nil.
//
// Callers should explicitly check for the possibility of a nil return, treat
// it as a failure of the system random number generator, and not continue.
func GenerateRandomKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
