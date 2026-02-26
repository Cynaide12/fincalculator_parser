package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string     `yaml:"env" env-default:"local" env-required:"true" env:"ENV"`
	LogFilePath string     `yaml:"log_file_path" env-default:".logs/fincalparser.log" env:"LOG_FILE_PATH"`
	DataDir     string     `yaml:"data_dir" env-default:"data_dir" env:"DATA_DIR"`
	HTTPServer  HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080" env:"HTTP_SERVER_ADDRESS"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s" env:"HTTP_SERVER_TIMEOUT"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s" env:"HTTP_SERVER_IDLE_TIMEOUT"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	//перекрываем значение из yaml переменными из env(чтобы можно было, например, в докере указать env)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}

	return &cfg
}
