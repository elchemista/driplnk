package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadJSONConfig(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	return nil
}
