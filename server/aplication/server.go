package aplication

import (
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/alejbv/SistemaFicherosRe/server/chord"
	"github.com/alejbv/SistemaFicherosRe/server/storage"
)

var (
	nodeFile   *chord.Node
	nodeTag    *chord.Node
	rsaPrivate string
	rsaPublic  string
)

func StartServer(network string, rsaPrivateKeyPath string, rsaPublicteKeyPath string) {
	var err error

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("El servicio de Sistema de Ficheros a empezado")

	rsaPrivate = rsaPrivateKeyPath
	rsaPublic = rsaPublicteKeyPath

	//Se crea el almacenamiento para los archivos
	fileDictionary, _ := storage.NewFileStorage(sha1.New)
	//Se crea un nodo chord para las etiquetas
	nodeFile, err = chord.DefaultNode("50050", fileDictionary)
	if err != nil {
		log.Fatalf("No se pudo crear el nodo para los archivos")
	}

	//Se crea el almacenamiento para los archivos
	tagDictionary, _ := storage.NewTagStorage(sha1.New)
	//Se crea un nodo chord para las etiquetas
	nodeTag, err = chord.DefaultNode("50051", tagDictionary)
	if err != nil {
		log.Fatalf("No se pudo crear el nodo para las etiquetas")
	}
	err = nodeFile.Start()

	if err != nil {
		log.Fatalf("No se pudo iniciar el nodo de los archivos")
	}

	err = nodeTag.Start()
	if err != nil {
		log.Fatalf("No se pudo iniciar el nodo de las etiquetas")
	}

	go StartTagService("tcp", "0.0.0.0:50052")
}
