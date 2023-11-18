package wal

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

type WAL interface {
	Append(cmd string) error
	Startup(m map[string]string) error
}

type WriteAheadLog struct {
	file *os.File
}

func New(filePath string) (*WriteAheadLog, error) {
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

	return &WriteAheadLog{file: file}, nil
}

func (w *WriteAheadLog) Append(cmd string) error {
	_, err := w.file.WriteString(cmd + "\n")
	if err != nil {
		return err
	}
	log.Info().Msgf("Appended: %s", cmd)
	return nil
}

func (w *WriteAheadLog) Startup(m map[string]string) error {
	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(w.file)
	log.Info().Msgf("Starting up... loading records to memory")

	// Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		log.Info().Msgf("Loading: %s", line)
		// Parse the command
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue // TODO: probably need to error, since there shouldn't be empty lines
		}

		// Check the command and validate arguments
		switch parts[0] {
		case "set":
			if len(parts) != 3 {
				continue
			}
			key := parts[1]
			val := parts[2]
			m[key] = val

		case "del":
			if len(parts) != 2 {
				fmt.Println("Invalid 'del' command. Usage: del {key}")
				continue
			}
			key := parts[1]
			delete(m, key)

		default:
			fmt.Println("Invalid command in file")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
