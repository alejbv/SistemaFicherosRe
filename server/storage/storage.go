package storage

type Storage interface {
	Set(string, []byte) error
	Get(string) ([]byte, error)
	Delete(string) error
	DeleteElemn(string, []string) error
	SetElem(string, []string) error
	Partition([]byte, []byte) (map[string][]byte, map[string][]byte, error)
	Extend(map[string][]byte) error
	Discard([]string) error
}
