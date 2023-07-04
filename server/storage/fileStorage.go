package storage

import (
	"hash"

	"github.com/alejbv/SistemaFicherosRe/server/utils"
)

type fileStorage struct {
	data map[string][]byte // Diccionario interno
	Hash func() hash.Hash  // Funcion Hash a usar
}

func NewFileStorage(hash func() hash.Hash) *fileStorage {
	return &fileStorage{
		data: make(map[string][]byte),
		Hash: hash,
	}
}

func (storage *fileStorage) Set(key string, file []byte) {
	storage.data[key] = file
}
func (storage *fileStorage) Get(key string) []byte {
	value, ok := storage.data[key]

	if !ok {
		return nil
	}

	return value
}
func (storage *fileStorage) Delete(key string) {
	delete(storage.data, key)
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
