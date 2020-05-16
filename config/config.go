package config

import (
	"os"
	"time"
)

// Config contains all configurations for this service
type Config struct {
	RC             *Redis
	RequestTimeout time.Duration // Request timeout in seconds. Defaults to 2 seconds
}

// GetConfig initializes the configuration for this
// service
// TODO : Get configuration from ENV variables
func GetConfig() *Config {
	return &Config{
		RC: &Redis{
			ConnectionType: "tcp",
			Host:           getEnv("REDIS_HOST", "127.0.0.1").(string),
			Port:           getEnv("REDIS_PORT", 6379).(int),
			Username:       getEnv("REDIS_USERNAME", "").(string),
			Password:       getEnv("REDIS_PASSWORD", "").(string),
			MaxIdle:        getEnv("REDIS_MAX_IDLE_CONNECTIONS", 3).(int),
			IdleTimeout:    getEnv("REDIS_IDLE_TIMEOUT", 240*time.Second).(time.Duration),
		},
		RequestTimeout: getEnv("REQUEST_TIMEOUT", 2*time.Second).(time.Duration),
	}
}

func getEnv(key string, fallback interface{}) interface{} {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
