package command

import (
	"bytes"
	"encoding/gob"
)

type Cmd struct {
	Cmd string
	Key string
	Val string
}

func New(cmd, key, val string) Cmd {
	return Cmd{cmd, key, val}
}

func (cmd *Cmd) GetVal() (string, error) {
	return cmd.Val, nil
}

func (cmd *Cmd) Serialize() ([]byte, error) {
	// Create a buffer to store the encoded data
	var buffer bytes.Buffer

	// Create an encoder
	encoder := gob.NewEncoder(&buffer)

	// Encode the struct to Go binary format
	err := encoder.Encode(cmd)
	if err != nil {
		return []byte{}, err
	}

	// Get the serialized bytes
	cmdBytes := buffer.Bytes()

	return cmdBytes, nil
}

func Deserialize(cmdBytes []byte) (Cmd, error) {
	// Create a buffer to read the encoded data
	buffer := bytes.NewBuffer(cmdBytes)

	gob.Register(Cmd{})

	// Create a decoder
	decoder := gob.NewDecoder(buffer)

	var cmd Cmd
	// Decode the bytes into the struct
	err := decoder.Decode(&cmd)
	if err != nil {
		return Cmd{}, err
	}

	return cmd, nil
}
