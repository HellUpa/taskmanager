package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env       string          `yaml:"env" env-default:"local"`
	Database  DatabaseConfig  `yaml:"db" env-required:"true"`
	GRPC      GRPCConfig      `yaml:"grpc"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
}
type DatabaseConfig struct {
	DBHost     string `yaml:"host"`
	DBPort     string `yaml:"port"`
	DBUser     string `yaml:"user"`
	DBPassword string `yaml:"password"`
	DBName     string `yaml:"name"`
}

type GRPCConfig struct {
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type TelemetryConfig struct {
	HealthCheckPort int `yaml:"port"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config read goes wrong: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
