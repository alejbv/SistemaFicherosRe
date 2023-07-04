package aplication

import (
	"fmt"
	"log"

	"github.com/alejbv/SistemaFicherosRe/server/chord"
)

var (
	node       *chord.Node
	rsaPrivate string
	rsaPublic  string
)

func StartServer(network string, rsaPrivateKeyPath string, rsaPublicteKeyPath string) {
	var err error

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("El servicio de Sistema de Ficheros a empezado")

	rsaPrivate = rsaPrivateKeyPath
	rsaPublic = rsaPublicteKeyPath

	//Se crea un nodo chord para los archvios
	node, err = chord.DefaultNode("50050")
	if err != nil {
		log.Fatalf("No se pudo crear el nodo chord")
	}
	err = node.Start()

	if err != nil {
		log.Fatalf("No se pudo iniciar el nodo de chord")
	}

	go StartTagService("tcp", "0.0.0.0:50051")
}
