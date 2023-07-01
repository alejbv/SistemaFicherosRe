package aplication

import (
	"context"
	"net"

	"github.com/alejbv/SistemaFicherosRe/service"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
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
		El primer paso es muy similar a la anterior. Es necesario recuperar todos los archivos que coinciden
		con las etiquetas de la query y pedirle al chord de los archivos su informacion. Por ultimo se reune toda la informacion y se devuelve

	*/
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
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				UnaryLoggingInterceptor,
				UnaryServerInterceptor,
			),
		), grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				StreamLoggingInterceptor,
				StreamServerInterceptor,
			),
		),
	)

	service.RegisterAplicationServer(s, &aplicationServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
