package tmpjson

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Write[T any](name string, v T) error {
	f := filepath.Join(os.TempDir(), name)
	return WriteJSONFile(f, v)
}

func Read[T any](name string, v *T) (*T, error) {
	f := filepath.Join(os.TempDir(), name)
	return ReadJSONFile(f, v)
}

func WriteJSONFile[T any](file string, v T) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(file, data, os.ModePerm); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func ReadJSONFile[T any](file string, v *T) (*T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", file, err)
	}

	return v, nil
}
