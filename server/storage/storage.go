package storage

type Storage interface {
	Set(string, []byte)
	Get(string) []byte
	Delete(string)
	Partition([]byte, []byte) (map[string][]byte, map[string][]byte, error)
	Extend(map[string][]byte) error
	Discard([]string) error
}
