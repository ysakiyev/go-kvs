package kvs

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"go-kvs/internal/server/wal"
	"go-kvs/pkg/kvs/command"
	"io"
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

	k := Kvs{
		index: index,
		wal:   wall,
	}

	err = k.Init()
	if err != nil {
		return nil, err
	}

	return &k, nil
}

func (k *Kvs) Init() error {
	var offset int64
	for {
		cmdBytes, newOffset, readErr := k.wal.Read(offset)
		if readErr != nil {
			if readErr == bufio.ErrFinalToken || readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				break
			}
			return readErr
		}

		// Create a buffer to read the encoded data
		buffer := bytes.NewBuffer(cmdBytes)

		// Create a decoder
		decoder := gob.NewDecoder(buffer)

		var cmd command.Cmd
		// Decode the bytes into the struct
		dErr := decoder.Decode(&cmd)
		if dErr != nil {
			if dErr == io.EOF {
				break
			}
			return dErr
		}

		switch cmd.Cmd {
		case "set":
			k.index[cmd.Key] = offset
		case "del":
			delete(k.index, cmd.Key)
		}

		offset = newOffset
	}

	return nil
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
	_, exists := k.index[key]
	if !exists {
		return fmt.Errorf("key doesn't exist")
	}

	cmd := command.New("del", key, "")
	cmdBytes, err := cmd.Serialize()
	if err != nil {
		return err
	}

	// append Cmd bytes to file and delete from index
	_, err = k.wal.Append(cmdBytes)
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

	cmdBytes, _, err := k.wal.Read(offset)
	if err != nil {
		return "", err
	}

	cmd, err := command.Deserialize(cmdBytes)
	if err != nil {
		return "", err
	}

	val, err := cmd.GetVal()
	if err != nil {
		return "", nil
	}

	return val, nil
}
