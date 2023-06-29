package storage

import (
	"encoding/json"
	"errors"
	"hash"

	"github.com/alejbv/SistemaFicherosRe/server/utils"
)

type fileStorage struct {
	data map[string][]byte // Diccionario interno
	Hash func() hash.Hash  // Funcion Hash a usar
}

func NewFileStorage(hash func() hash.Hash) (*fileStorage, error) {
	return &fileStorage{
		data: make(map[string][]byte),
		Hash: hash,
	}, nil
}

func (storage *fileStorage) Set(key string, file []byte) error {
	storage.data[key] = file
	return nil
}
func (storage *fileStorage) Get(key string) ([]byte, error) {
	value, ok := storage.data[key]

	if !ok {
		return nil, errors.New("llave no encontrada")
	}

	return value, nil
}
func (storage *fileStorage) Delete(key string) error {
	delete(storage.data, key)
	return nil
}
func (storage *fileStorage) DeleteElemn(key string, elemn []string) error {
	value, ok := storage.data[key]
	if !ok {
		return errors.New("llave no encontrada")
	}
	// La informacion codificada en el []bytes es un diccionario con la informacion de cada archivos
	// por lo que para trabajar en ella se quiere decodificar esa informacion
	var info FileEncoding
	err := json.Unmarshal(value, &info)
	if err != nil {
		return errors.New("Hubo un error tratando de decodificar la informacion de los archivos" + err.Error())
	}
	// Si al eliminar las etiquetas nos quedamos sin nada quiere decir que no hay ninguna etiqueta que
	// referencie a este fichero, por lo que se elimina del almacenamiento
	info.Tags = utils.DeleteElemnts(info.Tags, elemn)
	if len(info.Tags) == 0 {
		storage.Delete(key)
		return nil
	}

	// En otro caso, si aun quedan etiquetas se almacena otra vez
	convertInfo, err := json.MarshalIndent(&info, "", "\t")
	storage.data[key] = convertInfo
	return err
}
func (storage *fileStorage) SetElem(key string, file []string) error {
	value, ok := storage.data[key]
	if !ok {
		return errors.New("llave no encontrada")
	}
	// La informacion codificada en el []bytes es un diccionario con la informacion de cada archivos
	// por lo que para trabajar en ella se quiere decodificar esa informacion
	var info FileEncoding
	err := json.Unmarshal(value, &info)
	if err != nil {
		return errors.New("Hubo un error tratando de decodificar la informacion de los archivos" + err.Error())
	}
	info.Tags = utils.InsertWithOutDuplicates(info.Tags, file)
	convertInfo, err := json.MarshalIndent(&info, "", "\t")
	storage.data[key] = convertInfo
	return err
}
func (storage *fileStorage) Partition(L []byte, R []byte) (map[string][]byte, map[string][]byte, error) {
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
func (storage *fileStorage) Discard(files []string) error {
	for _, key := range files {
		delete(storage.data, key)
	}
	return nil
}
func (storage *fileStorage) Extend(data map[string][]byte) error {
	for key, value := range data {
		storage.data[key] = value
	}
	return nil
}
