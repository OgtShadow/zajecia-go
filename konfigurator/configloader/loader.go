package configloader

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrInvalidYAML        = errors.New("invalid yaml format")
)

// LoadFromYAML loads YAML config from path into a strongly typed struct passed as generic T.
func LoadFromYAML[T any](path string) (T, error) {
	var cfg T

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, fmt.Errorf("%w: %s", ErrConfigFileNotFound, path)
		}
		return cfg, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, fmt.Errorf("%w in %s: %v", ErrInvalidYAML, path, err)
	}

	return cfg, nil
}
