package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type UpstreamConfig struct {
    ID  string `yaml:"id"`
    URL string `yaml:"url"`
}

type Config struct {
    Server struct {
        Port int `yaml:"port"`
    } `yaml:"server"`
    Upstreams []UpstreamConfig `yaml:"upstreams"`
}

func LoadConfig(filepath string) (*Config, error) {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}