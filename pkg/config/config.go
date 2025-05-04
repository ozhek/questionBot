package config

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

var (
	once       sync.Once
	settings   *config.Config
	configFlag string
)

func init() {
	// Parse the config flag
	flag.StringVar(&configFlag, "config", "local", "Configuration file to use (local, production, etc.)")
	flag.Parse()

	// Initialize the configuration
	once.Do(func() {
		settings = config.New("default")
		settings.AddDriver(yaml.Driver)

		configPath := fmt.Sprintf("config/config_%s.yml", configFlag)
		if err := settings.LoadFiles(configPath); err != nil {
			log.Printf("Failed to load config file: %v. Using default values.", err)
		}
	})
}

// GetString retrieves a string value from the configuration or returns an empty string if settings is nil.
func GetString(key string) string {
	if settings == nil {
		return ""
	}
	return settings.String(key)
}

// GetInt retrieves an integer value from the configuration or returns 0 if settings is nil.
func GetInt(key string) int {
	if settings == nil {
		return 0
	}
	return settings.Int(key)
}

// GetInt32 retrieves an int32 value from the configuration or returns 0 if settings is nil.
func GetInt32(key string) int32 {
	if settings == nil {
		return 0
	}
	return int32(settings.Int(key))
}

// GetInt64 retrieves an int64 value from the configuration or returns 0 if settings is nil.
func GetInt64(key string) int64 {
	if settings == nil {
		return 0
	}
	return settings.Int64(key)
}

// GetFloat32 retrieves a float32 value from the configuration or returns 0.0 if settings is nil.
func GetFloat32(key string) float32 {
	if settings == nil {
		return 0.0
	}
	return float32(settings.Float(key))
}

// GetFloat64 retrieves a float64 value from the configuration or returns 0.0 if settings is nil.
func GetFloat64(key string) float64 {
	if settings == nil {
		return 0.0
	}
	return float64(settings.Float(key))
}

// GetBool retrieves a boolean value from the configuration or returns false if settings is nil.
func GetBool(key string) bool {
	if settings == nil {
		return false
	}
	return settings.Bool(key)
}
