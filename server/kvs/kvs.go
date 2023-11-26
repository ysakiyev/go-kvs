package kvs

import (
	"bufio"
	"fmt"
	"go-kvs/server/kvs/command"
	"go-kvs/server/wal"
)

type Kvs struct {
	index map[string]int64
	wal   wal.WAL
}

func New(fileName string) (*Kvs, error) {
	index := make(map[string]int64)
	wall, err := wal.New(fileName, index)
	if err != nil {
		return nil, err
	}

	var offset int64
	for {
		cmdBytes, rerr := wall.Read(offset)
		if rerr != nil {
			if rerr == bufio.ErrFinalToken {
				break
			}
			return nil, rerr
		}

		cmd, err := command.Deserialize(cmdBytes)
		if err != nil {
			return nil, err
		}

		switch cmd.Cmd {
		case "set":
			index[cmd.Key] = offset
		case "del":
			delete(index, cmd.Key)
		}

		offset += int64(len(cmdBytes)) + 1
	}

	return &Kvs{
		index: index,
		wal:   wall,
	}, nil
}

func (k *Kvs) Set(key, val string) error {
	// create Cmd object and serialize to bytes
	cmd := command.New("set", key, val)
	cmdBytes, err := cmd.Serialize()
	if err != nil {
		return err
	}

	// append Cmd bytes to file and store offset in memory index
	offset, err := k.wal.Append(cmdBytes)
	if err != nil {
		return err
	}
	k.index[key] = offset

	return nil
}

func (k *Kvs) Del(key string) error {
	cmd := command.New("del", key, "")
	bytes, err := cmd.Serialize()
	if err != nil {
		return err
	}
	_, err = k.wal.Append(bytes)
	if err != nil {
		return err
	}
	delete(k.index, key)

	return nil
}

func (k *Kvs) Get(key string) (string, error) {
	offset, exists := k.index[key]
	if !exists {
		return "", fmt.Errorf("key doesn't exist")
	}

	bytes, err := k.wal.Read(offset)
	if err != nil {
		return "", err
	}

	cmd, err := command.Deserialize(bytes)
	if err != nil {
		return "", err
	}

	val, err := cmd.GetVal()
	if err != nil {
		return "", nil
	}

	return val, nil
}
