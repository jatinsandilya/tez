package config

import "time"

// Redis contains the redis specific configuration
type Redis struct {
	ConnectionType string
	Host           string
	Port           int
	Username       string
	Password       string
	Name           string
	MaxIdle        int
	IdleTimeout    time.Duration
}
