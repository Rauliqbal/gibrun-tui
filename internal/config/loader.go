package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Ports    []int             `yaml:"ports"`
	Services map[string]string `yaml:"services"`
}

type Services map[string]Service

func Load(path string) (Services, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Services
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
