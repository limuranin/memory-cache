package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const EnvironmentPrefix string = "mc"

type ServerCfg struct {
	ListenAddress string `desc:"Server listen address" default:"127.0.0.1:8080" split_words:"true"`
}

type CacheCfg struct {
	CleaningInterval time.Duration `desc:"Cleaning cache interval" default:"30s" split_words:"true"`
}

type Config struct {
	Server *ServerCfg
	Cache  *CacheCfg
}

func Init() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(EnvironmentPrefix, cfg); err != nil {
		return nil, fmt.Errorf("environment configuration process error: %s", err)
	}

	return cfg, nil
}

func PrintHelp() error {
	return envconfig.Usage(EnvironmentPrefix, &Config{})
}
