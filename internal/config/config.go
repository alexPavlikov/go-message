package config

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env      string        `mapstructure:"env"`
	Timeout  time.Duration `mapstructure:"timeout"`
	LogLevel string        `mapstructure:"loglevel"`
	Server   Server        `mapstructure:"server"`
	Kafka    Server        `mapstructure:"kafka"`
	Topic    string        `mapstructure:"topic"`
	Postrges Postgres      `mapstructure:"postgres"`
}

type Server struct {
	Path string `mapstructure:"path"`
	Port int    `mapstructure:"port"`
}

type Postgres struct {
	Server   Server `mapstructure:"server"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DB       string `mapstructure:"name"`
}

func (s *Server) ToString() string {
	return fmt.Sprintf("%s:%d", s.Path, s.Port)
}

func Load() (*Config, error) {
	var cfg Config

	path, filename := fetchConfig()

	if path == "" || filename == "" {
		return &Config{}, errors.New("flag and system var empty")
	}

	if err := initViper(path, filename, &cfg); err != nil {
		return &Config{}, fmt.Errorf("failed to init viper: %w", err)
	}

	initLogger(cfg.LogLevel)

	slog.Info("load config complited", "config", cfg.Env)

	return &cfg, nil
}

func fetchConfig() (path string, filename string) {
	flag.StringVar(&path, "config_path", "", "path to config file")
	flag.StringVar(&filename, "config_file", "", "config file name")

	flag.Parse()

	if path == "" && filename == "" {
		path = os.Getenv("Config_PATH")
		filename = os.Getenv("Filename")
	}

	return path, filename
}

func initViper(path string, filename string, cfg *Config) error {
	viper.SetConfigName(filename)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal file to struct: %w", err)
	}

	return nil
}

func initLogger(loglevel string) {
	var log *slog.Logger
	switch loglevel {
	case "dev":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "prod":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	slog.SetDefault(log)
}
