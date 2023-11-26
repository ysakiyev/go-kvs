package wal

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"os"
)

type WAL interface {
	Append(cmd []byte) (int64, error)
	Read(offset int64) ([]byte, error)
}

type WriteAheadLog struct {
	file  *os.File
	index map[string]int64
}

func New(filePath string, index map[string]int64) (*WriteAheadLog, error) {
	// check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// create
		_, errCreate := os.Create(filePath)
		if errCreate != nil {
			return nil, errCreate
		}
		log.Info().Msg("Write-ahead log file created")
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &WriteAheadLog{file: file, index: index}, nil
}

func (w *WriteAheadLog) Append(cmd []byte) (int64, error) {
	offset, err := w.file.Seek(0, 1)
	if err != nil {
		return 0, err
	}

	cmd = append(cmd, '\n')

	_, err = w.file.Write(cmd)
	if err != nil {
		return 0, err
	}
	log.Info().Msgf("Appended: %s", cmd)
	return offset, nil
}

func (w *WriteAheadLog) Read(offset int64) ([]byte, error) {
	// Set the file cursor to the specified offset
	_, err := w.file.Seek(offset, 0)
	if err != nil {
		return []byte{}, err
	}

	// Create a buffered reader for efficient reading
	reader := bufio.NewReader(w.file)

	// Read until the end of the line
	lineBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return []byte{}, err
	}

	return lineBytes, nil
}
