package aplication

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/alejbv/SistemaFicherosRe/chord"
	"github.com/alejbv/SistemaFicherosRe/server/storage"
	"github.com/alejbv/SistemaFicherosRe/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type aplicationServer struct {
	service.UnimplementedAplicationServer
}

func (apServer *aplicationServer) AddFile(ctx context.Context, req *service.AddFileRequest) (*service.AddFileResponse, error) {
	/*
		Metodo para agregar un archivo al servidor.
		En la request se recibe el nombre, la extension y el tamaño en bytes del archivo ha subir
		Para luego obtener la fecha en la que se sube al servidor. Usando esos 4 elementos se computa un identificador
		unico para cada archivo. Luego  se añade el archivo al nodo chord de archivos usando SetFile.
		Por ultimo se recorren todas las etiquetas para a cada una agregarle que tiene un nuevo archivo
		Al final se devuelve un mensaje diciendo si hubo algun error
	*/

	// Se quiere obtener el tiempo actual para tenerlo en cuenta a la hora de hacer el identificador
	uploadDate := time.Now().String()

	//Se crea el identificador
	ident := fmt.Sprintf("%s%s%s%d%s", req.FileName, req.FileExtension, uploadDate, string(req.BytesSize), string(req.InfoToStorage))
	// El pbjetivo es serializar la informacion a almancer. Para eso primero se lleva a un formato en especifico de json
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
	chordRequest := &chord.SetRequest{
		Ident:       ident,
		File:        fileBytes,
		Replication: false,
	}
	// Se manda la request a almacenar. Si no hay error se sigue
	err := nodeFile.SetKey(ident, chordRequest)
	if err != nil {
		log.Error("Hubo un error almancenando el archivo" + err.Error())
		return &service.AddFileResponse{Response: err.Error()}, err
	}
	// Se prosigue almacenando en todas las etiquetas el identificador del archivo subido
	for _, tag := range req.Tags {
		// Se genera una request por cada etiqueta
		tagRequest := chord.SetRequest{Ident: tag, File: []byte(ident), Replication: false}
		err = nodeTag.SetKey(tag, &tagRequest)
		if err != nil {
			log.Errorf("Hubo un error almancenando la etiqueta: %s. %s", tag, err.Error())
			return &service.AddFileResponse{Response: err.Error()}, err
		}

	}
	return &service.AddFileResponse{Response: "Se almaceno el archivo sin problema"}, nil
}

func (apServer *aplicationServer) DeleteFile(ctx context.Context, req *service.DeleteFileRequest) (*service.DeleteFileResponse, error) {
	/*
		Metodo para eliminar archivos del servidor
		En la request se recibe las etiquetas para identificar los  archivos a eliminar. Para eso todos los archivos que
		contengan todas las etiquetas son eliminados.
		Primero se recupera la informacion de cada una de las etiquetas y se calcula la interseccion y luego se le manda al
		chord de los archivos que los elimine
		Luego se manda al chord de las etiquetas que elimine de cada etiqueta la informacion de los archivos eliminados
	*/

}

func (apServer *aplicationServer) ListFile(ctx context.Context, req *service.ListFileRequest) (*service.ListFileResponse, error) {
	/*
		Metodo para listar la informacion de los archivos en el servidor
		Se quiere obtener informacion de todos los archivos que se identifican por un conjunto de etiquetas.
		Para eso primero se va a recuperar la informacion de cada una de las etiquetes y solo interesa trabajar con los archivos
		que satisfacen completamente la Query.
	*/
	for _, tag := range req.QueryTags {
		nodeTag.GetKey(tag)
	}

}

func (apServer *aplicationServer) AddTags(ctx context.Context, req *service.AddTagsRequest) (*service.AddTagsResponse, error) {
	/*
		Metodo para agregar etiquetas a un conjunto

	*/

}
func (apServer *aplicationServer) DeleteTags(ctx context.Context, req *service.DeleteTagsRequest) (*service.DeleteTagsResponse, error) {
}

func StartTagService(network, address string) {
	log.Infof("Group Service Started")

	lis, err := net.Listen(network, address)

	if err != nil {
		log.Fatalf("Error al levantar el listener: %v", err)
	}
	s := grpc.NewServer(
	/*	grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				UnaryLoggingInterceptor,
				UnaryServerInterceptor,
			),
		), grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				StreamLoggingInterceptor,
				StreamServerInterceptor,
			),
		),*/
	)

	service.RegisterAplicationServer(s, &aplicationServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
