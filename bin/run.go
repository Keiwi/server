package main

import (
	"github.com/keiwi/server"
	"github.com/keiwi/utils/log"
	"github.com/keiwi/utils/log/handlers/cli"
	"github.com/keiwi/utils/log/handlers/file"
)

func main() {
	log.Log = log.NewLogger(log.DEBUG, []log.Reporter{
		cli.NewCli(),
		file.NewFile("./logs", "%date%_server.log"),
	})
	server.Start()
}
