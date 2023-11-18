package kvs

import (
	"fmt"
	"go-kvs/server/wal"
)

type Kvs struct {
	storage map[string]string
	wal     wal.WAL
}

func New() (*Kvs, error) {
	storage := make(map[string]string)
	wall, err := wal.New("wal.log", storage)
	if err != nil {
		return nil, err
	}

	err = wall.Startup()
	if err != nil {
		return nil, err
	}

	return &Kvs{
		storage: storage,
		wal:     wall,
	}, nil
}

func (k *Kvs) Get(key string) (string, error) {
	return k.storage[key], nil
}

func (k *Kvs) Set(key, val string) error {
	err := k.wal.Append(fmt.Sprintf("set %s %s", key, val))
	if err != nil {
		return err
	}
	k.storage[key] = val

	return nil
}

func (k *Kvs) Del(key string) error {
	err := k.wal.Append(fmt.Sprintf("del %s", key))
	if err != nil {
		return err
	}
	delete(k.storage, key)

	return nil
}
