package main

import (
	"github.com/alejbv/SistemaFicherosRe/server/aplication"
	"github.com/alejbv/SistemaFicherosRe/server/logging"
	log "github.com/sirupsen/logrus"
)

func main() {

	logging.SettingLogger(log.InfoLevel, "server")
	rsaPrivateKeyPath := "pv.pem"
	rsaPublicteKeyPath := "pub.pem"
	network := "tcp"
	aplication.StartServer(network, rsaPrivateKeyPath, rsaPublicteKeyPath)

	for {

	}
}
