package aplication

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	protoChord "github.com/alejbv/SistemaFicherosRe/chord"
	"github.com/alejbv/SistemaFicherosRe/server/storage"
	"github.com/alejbv/SistemaFicherosRe/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type aplicationServer struct {
	service.UnimplementedAplicationServer
}

//Ok
func (apServer *aplicationServer) AddFile(ctx context.Context, req *service.AddFileRequest) (*service.AddFileResponse, error) {
	/*
		Metodo para agregar un archivo al servidor.
		En la request se recibe el nombre, la extension y el tamaño en bytes del archivo ha subir
		Para luego obtener la fecha en la que se sube al servidor. Usando esos 4 elementos se computa un identificador
		unico para cada archivo. Luego  se añade el archivo al nodo chord de archivos usando SetFile.
		Por ultimo se recorren todas las etiquetas para a cada una agregarle que tiene un nuevo archivo
		Al final se devuelve un mensaje diciendo si hubo algun error
	*/
	// Se manda la request a almacenar. Si no hay error se sigue
	ident, chordRequest := CreateFileIdentifier(req)
	err := node.SetKey(ident, chordRequest)
	if err != nil {
		log.Error("Hubo un error almancenando el archivo" + err.Error())
		return &service.AddFileResponse{Response: err.Error()}, err
	}
	// Se prosigue almacenando en todas las etiquetas el identificador del archivo subido
	for _, tag := range req.Tags {
		// Se genera una request por cada etiqueta
		newReq, err := ManageTagAdd(node, tag, ident)
		if err != nil {
			return &service.AddFileResponse{Response: err.Error()}, err
		}
		// Se manda a realizar la request Set para las etiquetas
		node.SetKey(newReq.Ident, newReq)

	}
	return &service.AddFileResponse{Response: "Se almaceno el archivo sin problema"}, nil
}

//Ok
func (apServer *aplicationServer) ListFile(ctx context.Context, req *service.ListFileRequest) (*service.ListFileResponse, error) {
	/*
		Metodo para listar la informacion de los archivos en el servidor
		Se quiere obtener informacion de todos los archivos que sastisfagan  un conjunto de etiquetas.
		Para eso primero se va a recuperar la informacion de cada una de las etiquetes y solo interesa trabajar con los archivos
		que satisfacen completamente la Query.
	*/
	// Elemento donde se va a guardar la respuesta a esta request
	var sol map[string][]byte
	// realiza una Query para recuperar la informacion relevante para listar los archivos
	_, fileMaps, err := QueryTags(node, req.QueryTags)
	if err != nil {
		log.Error("Hubo un error recuperando la informacion" + err.Error())
		return &service.ListFileResponse{Response: nil}, err
	}

	// Se recorren todos los identificadores
	for key, value := range fileMaps {
		// Se comprueba si este elemento satisface la Query, lo que se determina si la
		// lista de etiquetas que mapea el archivo actual es igual a la de la Query
		if Equals(value, req.QueryTags) {
			fileInfo, err := node.GetKey(key)
			if err != nil {
				log.Errorf("Hubo un error recuperando la informacion del archivo %s: %s", key, err.Error())
				return &service.ListFileResponse{Response: nil}, err
			}
			sol[key] = fileInfo
		}

	}

	return &service.ListFileResponse{Response: sol}, nil
}

//Ok
func (apServer *aplicationServer) DeleteFile(ctx context.Context, req *service.DeleteFileRequest) (*service.DeleteFileResponse, error) {
	/*
		Metodo para eliminar archivos del servidor
		En la request se recibe las etiquetas para identificar los  archivos a eliminar. Para eso todos los archivos que
		contengan todas las etiquetas son eliminados.
		Primero se recupera la informacion de cada una de las etiquetas y se calcula la interseccion y luego se le manda al
		chord de los archivos que los elimine
		Luego se manda al chord de las etiquetas que elimine de cada etiqueta la informacion de los archivos eliminados
	*/
	// Primero las query
	var messageResponse string
	tagsMaps, fileMaps, err := QueryTags(node, req.DeleteTags)
	if err != nil {
		log.Error("Hubo un error recuperando la informacion" + err.Error())
		return &service.DeleteFileResponse{Response: err.Error()}, err
	}

	for fileKey, fileValue := range fileMaps {
		// Se comprueba si este elemento satisface la Query, lo que se determina si la
		// lista de etiquetas que mapea el archivo actual es igual a la de la Query
		if Equals(fileValue, req.DeleteTags) {
			err := node.DeleteKey(fileKey)
			if err != nil {
				messageResponse += fmt.Sprintf("Hubo un error eliminando el archivo %s: %s", fileKey, err.Error())
				log.Errorf("Hubo un error eliminando el archivo %s: %s", fileKey, err.Error())

				// Si no hubo problemas eliminando el archivo entonces se quiere  eliminar el archivo de las etiquetas
			} else {
				for tagsKey, tagsValue := range tagsMaps {
					tagsMaps[tagsKey] = DeleteElem(tagsValue, []string{fileKey})
				}
			}

		}

	}
	// Ahora entonces queda actualizar la informacion de las etiquetas en su almacenamiento
	for tagsKey, tagsValue := range tagsMaps {
		tagIdent := CreateTagIdentifier(tagsKey)
		fileBytes, _ := json.Marshal(tagsValue)
		setRequest := &protoChord.SetRequest{
			Ident:       tagIdent,
			File:        fileBytes,
			Replication: false,
		}
		node.SetKey(tagIdent, setRequest)
	}
	return &service.DeleteFileResponse{Response: messageResponse}, nil
}

func (apServer *aplicationServer) AddTags(ctx context.Context, req *service.AddTagsRequest) (*service.AddTagsResponse, error) {
	/*


	 */
	// Se crea esta variable para cuando se deserializen la informacion guardad de los archivos
	var tempDecoder storage.FileEncoding
	// Variable sobre la que se van a ir guardando todos los errores que han dado a lo largo de la ejecucion
	var messageResponse string
	// Variable que va a almacenar las nuevas etiquetas y los archivos que tienen
	var tagsMaps map[string][]string
	// Primero las query
	_, fileMaps, err := QueryTags(node, req.QueryTags)
	if err != nil {
		log.Error("Hubo un error recuperando la informacion" + err.Error())
		return &service.AddTagsResponse{Response: err.Error()}, err
	}
	// Se recorre la informacion del mapeo de archivos con las etiquetas de la Query que el satisface
	for fileKey, fileValue := range fileMaps {
		// Se comprueba si este elemento satisface la Query, lo que se determina si la
		// lista de etiquetas que mapea el archivo actual es igual a la de la Query
		if Equals(fileValue, req.QueryTags) {
			//Aqui se va a recuperar la informacion del archivo para modificarlo con las nuevas etiquetas
			currentFileInfo, err := node.GetKey(fileKey)
			if err != nil {
				messageResponse += fmt.Sprintf("Hubo un error recuperando la informacion del archivo %s: %s", fileKey, err.Error())
				log.Errorf("Hubo un error recuperando la informacion del archivo %s: %s", fileKey, err.Error())

				// Si no hubo problemas recuperando el archivo entonces se quiere  poder modificar del archivo
				// la informacion de las etiquetas
			} else {
				// Se deserializa la info en la variable FileEncoder creada al inicio
				json.Unmarshal(currentFileInfo, &tempDecoder)
				// Se insertan las nuevas etiquetas sin duplicados
				tempDecoder.Tags = InsertWithOutDuplicates(tempDecoder.Tags, req.AddTags)
				// Se vuelve a generar toda la SetRequest para el archivo para poder agregar el cambio
				// Para eso se vuelve a serializar la informacion
				fileBytes, _ := json.Marshal(tempDecoder)
				// Se crea una nueva request
				chordRequest := &protoChord.SetRequest{
					Ident:       fileKey,
					File:        fileBytes,
					Replication: false,
				}
				// Se manda la request
				err = node.SetKey(fileKey, chordRequest)
				if err != nil {
					messageResponse += fmt.Sprintf("Hubo un error almacenando la informacion del archivo %s: %s", fileKey, err.Error())
					log.Errorf("Hubo un error almacenando la informacion del archivo %s: %s", fileKey, err.Error())
				} else {
					// Ahora una vez almacenado el archivo actual sin problema es la hora de agregar la informacion
					// de las nuevas etiquetas
					for _, addTags := range req.AddTags {
						// Si no se a agregado la etiqueta actual, se agrega con el archivo actual como
						// unico elemento
						if tempValue, ok := tagsMaps[addTags]; !ok {
							tagsMaps[addTags] = []string{fileKey}
							// En otro caso se le agrega el nuevo elemento
						} else {
							tagsMaps[addTags] = InsertWithOutDuplicates(tempValue, []string{fileKey})
						}
					}
				}
			}

		}
	}
	// Ahora entonces queda almacenar la informacion de las nuevas etiquetas
	for tagsKey, tagsValue := range tagsMaps {
		tagIdent := CreateTagIdentifier(tagsKey)
		fileBytes, _ := json.Marshal(tagsValue)
		setRequest := &protoChord.SetRequest{
			Ident:       tagIdent,
			File:        fileBytes,
			Replication: false,
		}
		node.SetKey(tagIdent, setRequest)
	}

	return &service.AddTagsResponse{Response: messageResponse}, nil
}

func (apServer *aplicationServer) DeleteTags(ctx context.Context, req *service.DeleteTagsRequest) (*service.DeleteTagsResponse, error) {
	/*


	 */
	// Se crea esta variable para cuando se deserializen la informacion guardad de los archivos
	var tempDecoder storage.FileEncoding
	// Variable sobre la que se van a ir guardando todos los errores que han dado a lo largo de la ejecucion
	var messageResponse string
	// Variable que va a almacenar las  etiquetas y los archivos que quedan
	var deleteMaps map[string][]string
	// Primero las query
	tagMaps, fileMaps, err := QueryTags(node, req.QueryTags)
	if err != nil {
		log.Error("Hubo un error recuperando la informacion" + err.Error())
		return &service.DeleteTagsResponse{Response: err.Error()}, err
	}
	// Se recorre la informacion del mapeo de archivos con las etiquetas de la Query que el satisface
	for fileKey, fileValue := range fileMaps {
		// Se comprueba si este elemento satisface la Query, lo que se determina si la
		// lista de etiquetas que mapea el archivo actual es igual a la de la Query
		if Equals(fileValue, req.QueryTags) {
			//Aqui se va a recuperar la informacion del archivo para modificarlo con las nuevas etiquetas
			currentFileInfo, err := node.GetKey(fileKey)
			if err != nil {
				messageResponse += fmt.Sprintf("Hubo un error recuperando la informacion del archivo %s: %s", fileKey, err.Error())
				log.Errorf("Hubo un error recuperando la informacion del archivo %s: %s", fileKey, err.Error())

				// Si no hubo problemas recuperando el archivo entonces se quiere  poder modificar del archivo
				// la informacion de las etiquetas
			} else {
				// Se deserializa la info en la variable FileEncoder creada al inicio
				json.Unmarshal(currentFileInfo, &tempDecoder)
				// Se eliminan las etiquetas adecuadas
				tempDecoder.Tags = DeleteElem(tempDecoder.Tags, req.DeleteTags)
				// Se vuelve a generar toda la SetRequest para el archivo para poder agregar el cambio
				// Para eso se vuelve a serializar la informacion
				fileBytes, _ := json.Marshal(tempDecoder)
				// Se crea una nueva request
				chordRequest := &protoChord.SetRequest{
					Ident:       fileKey,
					File:        fileBytes,
					Replication: false,
				}
				// Se manda la request
				err = node.SetKey(fileKey, chordRequest)
				if err != nil {
					messageResponse += fmt.Sprintf("Hubo un error almacenando la informacion del archivo %s: %s", fileKey, err.Error())
					log.Errorf("Hubo un error almacenando la informacion del archivo %s: %s", fileKey, err.Error())
				} else {
					// Ahora una vez almacenado el archivo actual sin problema es la hora de agregar la informacion
					// de las nuevas etiquetas
					for _, delTags := range req.DeleteTags {
						// Si no se a agregado la etiqueta actual, se agrega con el archivo actual como
						// unico elemento
						if tempValue, ok := deleteMaps[delTags]; !ok {
							deleteMaps[delTags] = []string{fileKey}
							// En otro caso se le agrega el nuevo elemento
						} else {
							deleteMaps[delTags] = InsertWithOutDuplicates(tempValue, []string{fileKey})
						}
					}
				}
			}

		}
	}
	// Es necesario eliminar
	for tagsKey, tagsValue := range deleteMaps {
		deleteMaps[tagsKey] = DeleteElem(tagMaps[tagsKey], tagsValue)
	}
	// Ahora entonces queda almacenar la informacion de las nuevas etiquetas
	for tagsKey, tagsValue := range deleteMaps {
		tagIdent := CreateTagIdentifier(tagsKey)
		fileBytes, _ := json.Marshal(tagsValue)
		setRequest := &protoChord.SetRequest{
			Ident:       tagIdent,
			File:        fileBytes,
			Replication: false,
		}
		node.SetKey(tagIdent, setRequest)
	}

	return &service.DeleteTagsResponse{Response: messageResponse}, nil
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
