package config

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
)

type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	WriteTimeout   time.Duration `mapstructure:"write-timeout"`
	ReadTimeout    time.Duration `mapstructure:"read-timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle-timeout"`
	AllowedOrigins []string      `mapstructure:"allowed-origins"`
	LogLevel       string        `mapstructure:"log-level"`
}

func (cfg *ServerConfig) Validate() error {
	ip := net.ParseIP(cfg.Host)
	if ip == nil {
		return fmt.Errorf("invalid host: %v", cfg.Host)
	}

	if cfg.Port < 0 || cfg.Port > 65535 {
		return errors.New("invalid port")
	}

	if cfg.WriteTimeout < 0 {
		return errors.New("write timeout cannot be negative")
	}

	if cfg.ReadTimeout < 0 {
		return errors.New("read timeout cannot be negative")
	}

	if cfg.IdleTimeout < 0 {
		return errors.New("idle timeout cannot be negative")
	}

	return nil
}

func (cfg *ServerConfig) ValidateServerLogLevel() error {
	// If log level is not set, we don't need to validate it, a default value will be used in service
	if cfg.LogLevel == "" {
		return nil
	}

	if parsedLevel, err := zerolog.ParseLevel(cfg.LogLevel); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	} else if parsedLevel < zerolog.DebugLevel || parsedLevel > zerolog.FatalLevel {
		return fmt.Errorf("only log levels from debug to fatal are supported")
	}
	return nil
}