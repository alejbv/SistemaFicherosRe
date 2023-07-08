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

func (apServer *aplicationServer) AddFile(ctx context.Context, req *service.AddFileRequest) (*service.AddFileResponse, error) {
	/*
		Metodo para agregar un archivo al servidor.
		En la request se recibe el nombre, la extension y el tamaño en bytes del archivo a subir
		Para luego obtener la fecha en la que se sube al servidor.
		Usando esos 4 elementos se computa  un identificador unico para cada archivo.
		Luego  se añade el archivo al chord codificado con toda la informacion referente al archivo.
		Por ultimo se recorren todas las etiquetas para a cada una agregarle que tiene
		un nuevo archivo.
		Al final se devuelve un mensaje diciendo si hubo algun error
	*/
	// Se manda la informacion recibida de la request add para obtener el identificador correspondiente
	// asi como codificar la informacion de una forma adecuada para almacenarla en el chord
	ident, chordRequest := ProcessFile(req)

	// Una vez procesado los datos se manda a almacenar en el chord
	err := node.SetKey(ident, chordRequest)
	if err != nil {
		log.Error("Hubo un error almancenando el archivo" + err.Error())
		return &service.AddFileResponse{Response: err.Error()}, err
	}
	// Se prosigue almacenando en todas las etiquetas el identificador del archivo subido
	for _, tag := range req.Tags {
		// Se genera una request por cada etiqueta
		newReq, err := ManageTagAdd(node, tag, []string{ident})
		if err != nil {
			return &service.AddFileResponse{Response: err.Error()}, err
		}
		// Se manda a realizar la request Set para las etiquetas
		node.SetKey(newReq.Ident, newReq)

	}
	return &service.AddFileResponse{Response: fmt.Sprintf("Se almaceno el archivo %s.%s sin problema", req.FileName, req.FileExtension)}, nil
}

func (apServer *aplicationServer) ListFile(ctx context.Context, req *service.ListFileRequest) (*service.ListFileResponse, error) {
	/*
		Metodo para listar la informacion de los archivos en el servidor
		Se quiere obtener informacion de todos los archivos que sastisfagan  un conjunto de etiquetas.
		Para eso primero se va a recuperar la informacion de cada una de las etiquetas y solo interesa trabajar con los archivos
		que satisfacen completamente la Query.
	*/
	// Elemento donde se va a guardar la respuesta a esta request
	sol := make(map[string][]byte)
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
		if Contains(value, req.QueryTags) {
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

func (apServer *aplicationServer) DeleteFile(ctx context.Context, req *service.DeleteFileRequest) (*service.DeleteFileResponse, error) {
	/*
		Metodo para eliminar archivos del servidor
		En la request se recibe las etiquetas para identificar los  archivos a eliminar. Para eso todos los archivos que
		satisfagan la Query van a ser eliminados.
		Primero se recupera la informacion de cada una de las etiquetas y se calcula la interseccion y luego se le manda al
		chord de los archivos que los elimine
		Luego se manda al chord de las etiquetas que elimine de cada etiqueta la informacion de los archivos eliminados
	*/
	// Primero las query

	// String donde se van a almacenar los posibles error
	var messageResponse string
	// Slice de FileEncoding donde se va a deserializar la informacion necesaria
	var fileDecod storage.FileEncoding
	// Variable donde se van a almacenar para cada etiqueta todos los archivos a eliminar
	tagsMaps := make(map[string][]string)
	_, fileMaps, err := QueryTags(node, req.DeleteTags)
	if err != nil {
		log.Error("Hubo un error recuperando la informacion" + err.Error())
		return &service.DeleteFileResponse{Response: err.Error()}, err
	}

	for fileKey, fileValue := range fileMaps {
		// Se comprueba si este elemento satisface la Query, lo que se determina si la
		// lista de etiquetas que mapea el archivo actual es igual a la de la Query
		if Contains(fileValue, req.DeleteTags) {
			// Primero es necesario recuperar la informacion del archivo para poder usar sus etiquetas
			val, err := node.GetKey(fileKey)

			if err != nil {
				messageResponse += fmt.Sprintf("Hubo un error recuperando la informacion del archivo con identificador %s: %s", fileKey, err.Error())
				log.Errorf("Hubo un error recuperando la informacion del archivo con identificador %s: %s", fileKey, err.Error())
				// Si no hubo problemas recuperando el archivo entonces se quiere  eliminar el archivo de las etiquetas
			} else {
				err = node.DeleteKey(fileKey)
				if err != nil {
					messageResponse += fmt.Sprintf("Hubo un error eliminando el archivo con identificador %s: %s", fileKey, err.Error())
					log.Errorf("Hubo un error eliminando el archivo con identificador %s: %s", fileKey, err.Error())
					// Si no hubo problemas eliminando el archivo entonces se quiere  eliminar el archivo de las etiquetas
				} else {
					// Se decodifica la informacion para poder tener acceso a todas las etiquetas del
					// archivo actual
					json.Unmarshal(val, &fileDecod)
					// Se recorren todas las etiquetas del archivo actual y por cada una de ellas se le
					// agrega el archivo actual como un elemento a eliminar
					for _, currentFileTag := range fileDecod.Tags {
						if deleteFiles, ok := tagsMaps[currentFileTag]; !ok {
							// De no haber un elemento con la etiqueta actual se crea y se le añada el archivo
							tagsMaps[currentFileTag] = []string{fileKey}
						} else {
							tagsMaps[currentFileTag] = append(deleteFiles, fileKey)
						}
					}
				}
			}

		}

	}
	// Ahora entonces queda actualizar la informacion de las etiquetas en su almacenamiento
	// Se recorren todas las etiquetas con archivos a eliminar
	// Variable para almacenar temporalmente los archivos a eliminar por etiqueta
	var temp []string
	for tagsKey, tagsValue := range tagsMaps {
		tagIdent := CreateTagIdentifier(tagsKey)
		// Se recupera la informacion de la actual
		value, err := node.GetKey(tagIdent)
		if err != nil {
			messageResponse += fmt.Sprintf("Hubo un error para recuperar la informacion de la etiqueta %s :%s", tagsKey, err.Error())
			log.Errorf("Hubo un error para recuperar la informacion de la etiqueta %s :%s", tagsKey, err.Error())
			// Si no hubo problemas recuperando la informacion
		} else {
			// Se decodifica la informacion de la etiqueta para obtener todos los archivos que tiene
			json.Unmarshal(value, &temp)
			// Se elimina del total, el subconjunto que fue eliminado previamente
			tagsValue = DeleteElem(temp, tagsValue)
			// Si todavia quedan archivos se almancenan
			if len(tagsValue) > 0 {
				fileBytes, _ := json.Marshal(tagsValue)
				setRequest := &protoChord.SetRequest{
					Ident:       tagIdent,
					File:        fileBytes,
					Replication: false,
				}
				node.SetKey(tagIdent, setRequest)
				// En otro caso se manda a eliminar la etiqueta del Almacenamiento
			} else {
				node.DeleteKey(tagIdent)
			}

		}

	}
	return &service.DeleteFileResponse{Response: messageResponse}, nil
}

func (apServer *aplicationServer) AddTags(ctx context.Context, req *service.AddTagsRequest) (*service.AddTagsResponse, error) {
	/*
		Metodo para agregarle a los archivos que satisfacen la Query un conjunto nuevo de etiquetas.
		Primero se recupera toda la informacion de cada una de las etiquetas de la Query
		(Por cada etiqueta que archivos apunta ella) y de ahi se determinan todos los archivos
		que la cumplen. Por cada uno de ellos se agregan las nuevas etiquetas asegurando que no
		ocurran duplicados. Despues se manda a agregar a esas etiquetas todos los archivos que
		cumplieron la Query.

	*/
	// Se crea esta variable para cuando se deserializen la informacion guardad de los archivos
	var tempDecoder storage.FileEncoding
	// Variable sobre la que se van a ir guardando todos los errores que han dado a lo largo de la ejecucion
	var messageResponse string
	// Variable para almacenar todos los archivos que van a almacenar cada etiqueta
	var filesToAdd []string
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
		if Contains(fileValue, req.QueryTags) {
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

					filesToAdd = append(filesToAdd, fileKey)
				}
			}

		}
	}
	// Ahora entonces queda almacenar la informacion de las nuevas etiquetas
	for _, tagsValue := range req.AddTags {
		addReq, _ := ManageTagAdd(node, tagsValue, filesToAdd)
		node.SetKey(addReq.Ident, addReq)
	}

	return &service.AddTagsResponse{Response: messageResponse}, nil
}

func (apServer *aplicationServer) DeleteTags(ctx context.Context, req *service.DeleteTagsRequest) (*service.DeleteTagsResponse, error) {
	/*
		Metodo para remover de  los archivos que satisfacen la Query un conjunto  de etiquetas.
		Primero se recupera toda la informacion de cada una de las etiquetas de la Query
		(Por cada etiqueta que archivos apunta ella) y de ahi se determinan todos los archivos
		que la cumplen. Por cada uno de ellos se remuenven las  etiquetas.
		Despues se recorren las etiquetas que quedaron y por cada una se remueven los archivos que
		satisfacieron la Query
	*/
	// Se crea esta variable para cuando se deserializen la informacion guardad de los archivos
	var tempDecoder storage.FileEncoding
	// Variable sobre la que se van a ir guardando todos los errores que han dado a lo largo de la ejecucion
	var messageResponse string
	// Variable que va a almacenar las  etiquetas y los archivos que quedan
	deleteMaps := make(map[string][]string)
	// Variable que va a almacenar los archivos a eliminar
	var deleteFiles []string
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
		if Contains(fileValue, req.QueryTags) {
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
					// de los archivos que se debe eliminar de las etiquetas
					deleteFiles = append(deleteFiles, fileKey)
				}
			}

		}
	}
	// Es necesario eliminar
	for _, tagsValue := range req.DeleteTags {
		// Si esta etiqueta esta contenida dentro de los archivos recuperados entonces se eliminan directamente
		if tagInfo, ok := tagMaps[tagsValue]; ok {
			deleteMaps[tagsValue] = DeleteElem(tagInfo, deleteFiles)
			// En otro caso es necesario recuperar la informacion
		} else {
			// Variable temporal donde se va a almacenar la informacion a decodificar
			var temp []string
			// Se recupera la informacion de esa etiqueta
			value, err := node.GetKey(CreateTagIdentifier(tagsValue))
			if err != nil {
				messageResponse += fmt.Sprintf("Hubo un error recuperando la informacion de la etiqueta %s: %s", tagsValue, err.Error())
				log.Errorf("Hubo un error recuperando la informacion de la etiqueta %s: %s", tagsValue, err.Error())
				// De no haber error
			} else {
				// Se decodifica la informacion recuperada
				json.Unmarshal(value, &temp)
				deleteMaps[tagsValue] = DeleteElem(temp, deleteFiles)
			}
		}

	}
	// Ahora entonces queda almacenar la informacion de las nuevas etiquetas
	for tagsKey, tagsValue := range deleteMaps {
		tagIdent := CreateTagIdentifier(tagsKey)
		// Si tiene al menos algun elemento se almacena
		if len(tagsValue) > 0 {
			fileBytes, _ := json.Marshal(tagsValue)
			setRequest := &protoChord.SetRequest{
				Ident:       tagIdent,
				File:        fileBytes,
				Replication: false,
			}
			node.SetKey(tagIdent, setRequest)
		} else {
			// Si el elemento actual no tienen ningun archivo para agregar se elimina directamente
			node.DeleteKey(tagIdent)
		}
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
