package config

import (
	"time"

	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Server struct {
		ListenAddr   string        `yaml:"listen_addr"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
	} `yaml:"server"`

	LoadBalancer struct {
		Algorithm           string        `yaml:"algorithm"`
		HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	} `yaml:"load_balancer"`

	BackendServers []string `yaml:"backend_servers"`

	TLS struct {
		Enabled  bool   `yaml:"enabled"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	Metrics struct {
		Enabled bool `yaml:"enabled"`
		Port    int  `yaml:"port"`
	} `yaml:"metrics"`
}

func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}