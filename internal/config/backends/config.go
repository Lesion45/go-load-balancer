package backends

import (
	"os"

	"gopkg.in/yaml.v3"
)

type BackendConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Burst   int    `yaml:"burst"`
	Rate    int    `yaml:"rate"`
}

type Config struct {
	Backends []BackendConfig `yaml:"backends"`
}

// MustLoad loads the backend configuration from a YAML file.
// It panics if the file cannot be read or if the configuration is invalid.
func MustLoad() *Config {
	data, err := os.ReadFile("/app/backends/backends.yaml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
