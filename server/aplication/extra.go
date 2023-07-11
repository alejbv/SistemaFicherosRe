package aplication

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	protoChord "github.com/alejbv/SistemaFicherosRe/chord"
	serverChord "github.com/alejbv/SistemaFicherosRe/server/chord"
	"github.com/alejbv/SistemaFicherosRe/server/storage"
	"github.com/alejbv/SistemaFicherosRe/server/utils"
	"github.com/alejbv/SistemaFicherosRe/service"
	log "github.com/sirupsen/logrus"
)

func ProcessFile(req *service.AddFileRequest) (string, *protoChord.SetRequest) {
	// Se quiere obtener el tiempo actual para tenerlo en cuenta a la hora de hacer el identificador
	uploadDate := time.Now().String()
	temp := strings.Split(uploadDate, " ")
	uploadDate = temp[0] + " " + temp[1]

	//Se crea el identificador usando la informacion propia del archivo y la fecha de subida al servidor
	ident := fmt.Sprintf("File.%s%s%s%s%s", req.FileName, req.FileExtension, uploadDate, string(req.BytesSize), string(req.InfoToStorage))
	// El objetivo es serializar la informacion a almancer. Para eso primero se lleva a un formato en especifico de json
	// para luego llevarlo a bytes
	fileInfo := &storage.FileEncoding{
		FileName:      req.FileName,
		FileExtension: req.FileExtension,
		FileInfo:      req.InfoToStorage,
		UploadDate:    uploadDate,
		Size:          int(req.BytesSize),
		Tags:          req.Tags,
	}
	// Se codifica a bytes
	fileBytes, _ := json.Marshal(fileInfo)
	chordRequest := &protoChord.SetRequest{
		Ident:       ident,
		File:        fileBytes,
		Replication: false,
	}
	return ident, chordRequest
}

func CreateTagIdentifier(tag string) string {
	return fmt.Sprintf("Tag.%s", tag)
}

// Metodo para manegar la adiccion de una etiqueta a un nodo chord
func ManageTagAdd(node *serverChord.Node, tag string, fileIdent []string) (*protoChord.SetRequest, error) {
	/*
		Cuando se agrega un nuevo archivo al servicio Chord tambien se debe almacenar la informacion de la
		etiqueta, guardandose que una determinada etiqueta referencia a un archivo, Pero en esa operacion
		puede pasar una de dos cosas, Que el archivo no existe, o que si existe. En el primere caso no ocurre nada
		se crea la etiqueta y ya, pero en el segundo es necesario hacer un procedimiento para agregar la nueva etiqueta
	*/

	// primero se quiere obtener el identificador unico para una etiqueta
	tagIdent := CreateTagIdentifier(tag)
	// Luego se quiere comprobar si existe en el servicio dicha etiqueta
	file, err := node.GetKey(tagIdent)
	// Como se menciono antes, o el archivo existe o no. Para eso se
	// Hay un error y no es que el archivo sea nulo
	if err != nil {
		log.Errorf("Hubo un error a la hora de recuperar la informacion de la etiqueta %s: %s", tag, err.Error())
		return nil, err
	}

	var tagInfo []string
	// En caso de que el archivo sea distinto de nil, implica de que tiene informacion, por lo tanto
	// se extrae dicha informacion y se almacena en tagInfo
	if file != nil {
		json.Unmarshal(file, &tagInfo)
	}
	// Se inserta la nueva informacion y luego se codifica a bytes
	newFileInfo := InsertWithOutDuplicates(tagInfo, fileIdent)
	newCodeInfo, _ := json.Marshal(&newFileInfo)
	return &protoChord.SetRequest{
		Ident:       tagIdent,
		File:        newCodeInfo,
		Replication: false,
	}, nil
}

// Metodo para recuperar todos los archivos
func QueryTags(node *serverChord.Node, queryTags []string) (map[string][]string, map[string][]string, error) {
	// Se crea un diccionario que va a tener las etiquetas mapeando a todos los archivos que almacena
	tagsMap := make(map[string][]string)
	// Se crea un diccionario que va a tener por cada archivo todas las etiquetas que de la query que lo tienen
	filesMap := make(map[string][]string)
	// Slice temporal para deserializar la informacion
	var temp []string
	// Se recorre todas las etiquetas
	for _, tag := range queryTags {
		tagIdent := CreateTagIdentifier(tag)
		val, err := node.GetKey(tagIdent)
		if err != nil {
			log.Errorf("Hubo un error a la hora de recuperar la informacion de la etiqueta %s: %s", tag, err.Error())
			return nil, nil, err
		}
		// Se deserializa la informacion
		json.Unmarshal(val, &temp)
		// Se almacena para etiqueta actual su informacion relevante
		tagsMap[tag] = temp
		// Se va a recorrer cada uno de los archivos de la etiqueta actual
		for _, file := range temp {
			// Si el archivo no esta almacenado, se crea una nueva lista con la etiqueta actual
			if info, ok := filesMap[file]; !ok {
				filesMap[file] = []string{tag}
				// En cambio si existe se agrega el nuevo elemento
			} else {
				filesMap[file] = append(info, tag)
			}
		}
	}
	return tagsMap, filesMap, nil
}

// Metodo para comprobar si dos listas de strings son iguales
func Contains(first, second []string) bool {
	m := make(map[string]bool)

	// Se añaden todos los elementos de la primera lista al mapa
	for _, elem := range first {
		m[elem] = true
	}
	// Para que las dos listas sean iguales todos los elementos de la segunda deben estar en el mapa.
	// Si hay al menos un elemento de second que no este en m, entonces no son iguales
	for _, elem := range second {
		if _, ok := m[elem]; !ok {
			return false
		}
	}
	return true
}

// Metodo para eliminar de una lista de string un subconjunto de esta
func DeleteElem(elemList, elemDeletes []string) []string {
	m := make(map[string]bool)
	for _, elem := range elemList {
		m[elem] = true
	}
	for _, elemDelete := range elemDeletes {
		delete(m, elemDelete)
	}
	return utils.Keys(m)
}
func InsertWithOutDuplicates(first, second []string) []string {
	m := make(map[string]bool)

	for _, v := range first {
		m[v] = true
	}
	for _, v := range second {
		m[v] = true
	}

	return utils.Keys(m)
}
