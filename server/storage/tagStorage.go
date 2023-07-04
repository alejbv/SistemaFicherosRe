package storage

import (
	"encoding/json"
	"errors"
	"hash"

	"github.com/alejbv/SistemaFicherosRe/server/utils"
	log "github.com/sirupsen/logrus"
)

type tagStorage struct {
	data map[string][]byte // Diccionario interno
	Hash func() hash.Hash  // Funcion Hash a usar
}

func NewTagStorage(hash func() hash.Hash) (*tagStorage, error) {
	return &tagStorage{
		data: make(map[string][]byte),
		Hash: hash,
	}, nil
}
func (storage *tagStorage) Set(key string, file []byte) error {
	// Variable local que va a almacenar todos los archivos que tiene la etiqueta en key
	var tempStorage []string
	// Se comprueba si esta el valor de la etiqueta. De no estar se crea un nuevo arreglo con un
	// unico elemento, el nuevo y se lleva a binario para luego almacenarlo
	if val, ok := storage.data[key]; !ok {
		tempStorage = append(tempStorage, string(file))
		toBytes, err := json.Marshal(&tempStorage)
		if err != nil {
			log.Errorf("No se pudo almacenar el archvio: %s. %s", string(file), err.Error())
			return err
		}
		storage.data[key] = toBytes
		// De existir esa etiqueta se almacena temporalmente todos los archivos que tienene y se agrega el nuevo
		// siempre que no hayan duplicados
	} else {
		json.Unmarshal(val, &tempStorage)
		result := utils.InsertWithOutDuplicates(tempStorage, []string{string(file)})
		toBytes, _ := json.Marshal(&result)
		storage.data[key] = toBytes
	}
	return nil
}
func (storage *tagStorage) Get(key string) ([]byte, error) {
	value, ok := storage.data[key]

	if !ok {
		return nil, errors.New("llave no encontrada")
	}

	return value, nil
}
func (storage *tagStorage) Delete(key string) error {
	delete(storage.data, key)
	return nil
}
func (storage *tagStorage) DeleteElemn(key string, elemn []string) error {
	value, ok := storage.data[key]
	if !ok {
		return errors.New("llave no encontrada")
	}
	// La informacion codificada en el []bytes es una lista con la informacion de todos los archivos
	// que esta etiqueta referencia, por lo que para trabajar en ella se quiere decodificar esa informacion
	var info []string
	err := json.Unmarshal(value, &info)
	if err != nil {
		return errors.New("Hubo un error tratando de decodificar la informacion de los archivos" + err.Error())
	}
	// Si al eliminar los archivos nos quedamos sin nada quiere decir que no hay ningun archivo que
	// use a esta etiqueta, por lo que se elimina del almacenamiento
	info = utils.DeleteElemnts(info, elemn)
	if len(info) == 0 {
		storage.Delete(key)
		return nil
	}

	// En otro caso, si aun quedan archivos se almacena otra vez
	convertInfo, err := json.MarshalIndent(&info, "", "\t")
	storage.data[key] = convertInfo
	return err
}
func (storage *tagStorage) SetElem(key string, file []string) error {
	value, ok := storage.data[key]
	if !ok {
		return errors.New("llave no encontrada")
	}
	// La informacion codificada en el []bytes es un diccionario con la informacion de cada archivos
	// por lo que para trabajar en ella se quiere decodificar esa informacion
	var info []string
	err := json.Unmarshal(value, &info)
	if err != nil {
		return errors.New("Hubo un error tratando de decodificar la informacion de los archivos" + err.Error())
	}
	info = utils.InsertWithOutDuplicates(info, file)
	convertInfo, err := json.MarshalIndent(&info, "", "\t")
	storage.data[key] = convertInfo
	return err
}
func (storage *tagStorage) Partition(L []byte, R []byte) (map[string][]byte, map[string][]byte, error) {
	in := make(map[string][]byte)
	out := make(map[string][]byte)

	if utils.Equals(L, R) {
		return storage.data, out, nil
	}

	for key, value := range storage.data {
		if between, err := utils.KeyBetween(key, storage.Hash, L, R); between && err == nil {
			in[key] = value
		} else if err == nil {
			out[key] = value
		} else {
			return nil, nil, err
		}
	}

	return in, out, nil
}
func (storage *tagStorage) Discard(files []string) error {
	for _, key := range files {
		delete(storage.data, key)
	}
	return nil
}
func (storage *tagStorage) Extend(data map[string][]byte) error {
	for key, value := range data {
		storage.data[key] = value
	}
	return nil
}
