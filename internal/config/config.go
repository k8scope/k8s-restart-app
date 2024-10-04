package config

import (
	"os"

	"github.com/k8scope/k8s-restart-app/internal/k8s"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []k8s.KindNamespaceName `json:"services"`
}

// ReadConfigFile reads a yaml file and returns a Config struct
func ReadConfigFile(path string) (*Config, error) {
	bts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(bts, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
