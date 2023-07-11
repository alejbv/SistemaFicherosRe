package client

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/alejbv/SistemaFicherosRe/server/storage"
	"github.com/alejbv/SistemaFicherosRe/service"
	log "github.com/sirupsen/logrus"
)

func ClientAddFile(client service.AplicationClient, filePath string, tags []string) error {

	chain := strings.Split(filePath, "/")
	tempName := strings.Split(chain[len(chain)-1], ".")
	fileName := tempName[0]
	fileExtension := tempName[1]

	log.Printf("Se va a subir el archivo %s con etiquetas\n %s", fileName, tags)

	fileInfo, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorf("Error mientras se cargaba el archivo:\n" + err.Error())
		return errors.New("Error mientras se cargaba el archivo:\n" + err.Error())
	}
	req := &service.AddFileRequest{
		FileName:      fileName,
		FileExtension: fileExtension,
		InfoToStorage: fileInfo,
		BytesSize:     int32(len(fileInfo)),
		Tags:          tags,
	}
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.AddFile(ctx, req)
	if err != nil {
		log.Errorf("Error mientras se subia el archivo:\n" + err.Error())
		return errors.New("Error mientras se subia el archivo:\n" + err.Error())

	}
	return nil
}
func ClientDeleteFile(client service.AplicationClient, tags []string) error {

	log.Printf("Se van a eliminar todos los archivos  con etiquetas\n %s", tags)
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.DeleteFile(ctx, &service.DeleteFileRequest{DeleteTags: tags})
	if err != nil {
		log.Errorf("Error mientras se eliminaban los archivos:\n" + err.Error())
		return errors.New("Error mientras se eliminaban los archivos:\n" + err.Error())

	}
	return nil
}
func ClientListFiles(client service.AplicationClient, tags []string) error {

	log.Infof("Se va a obtener la informacion de los archivos con etiquetas\n %s", tags)
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.ListFile(ctx, &service.ListFileRequest{QueryTags: tags})
	if err != nil {
		log.Errorf("Error mientras se obtenia la informacion:\n" + err.Error())
		return errors.New("Error mientras se obtenia la informacion:\n" + err.Error())

	}
	info := res.Response

	// Imprimir encabezado de la tabla
	log.Printf("| %20s | %20s | %20s | %20s | %20s |\n", "Nombre del Archivo", "Extension ", "Tamaño", "Fecha de Subida", "Etiquetas")
	str := "-"
	repetitions := 100
	result := strings.Repeat(str, repetitions)
	log.Println(result)
	// Imprimir datos de la tabla
	for _, value := range info {
		var fileDecod storage.FileEncoding
		json.Unmarshal(value, &fileDecod)
		log.Printf("| %20s | %20s | %20d | %20s | %20s |\n", fileDecod.FileName, fileDecod.FileExtension, fileDecod.Size, fileDecod.UploadDate, strings.Join(fileDecod.Tags, ","))
	}

	return nil

}
func ClientAddTags(client service.AplicationClient, queryTags, addTags []string) error {

	log.Printf("Se van a agregar a todos los archivos  con etiquetas\n %s\n Las etiquetas:%s", queryTags, addTags)
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.AddTags(ctx, &service.AddTagsRequest{QueryTags: queryTags, AddTags: addTags})
	if err != nil {
		log.Errorf("Error mientras se agregaban las etiquetas:\n" + err.Error())
		return errors.New("Error mientras se agregaban las etiquetas:\n" + err.Error())

	}
	return nil
}

func ClientDeleteTags(client service.AplicationClient, queryTags, deleteTags []string) error {

	log.Printf("Se van a eliminar de todos los archivos  con etiquetas\n %s\n Las etiquetas:%s", queryTags, deleteTags)
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.DeleteTags(ctx, &service.DeleteTagsRequest{QueryTags: queryTags, DeleteTags: deleteTags})
	if err != nil {
		log.Errorf("Error mientras se eliminaban las etiquetas:\n" + err.Error())
		return errors.New("Error mientras se eliminaban las etiquetas:\n" + err.Error())

	}
	return nil
}
