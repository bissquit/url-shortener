package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileObserver struct {
	filepath string
}

func NewFileObserver(filepath string) *FileObserver {
	return &FileObserver{
		filepath: filepath,
	}
}

func (f *FileObserver) Notify(event Event) error {
	file, err := os.OpenFile(f.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	return nil
}
