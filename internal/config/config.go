package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v8"
)

type Config struct {
	Env  string `env:"ENV,required"`
	Host string `env:"HOST,required"`
	Port string `env:"PORT,required"`
	LoadBalancerConfig
	HealthCheckConfig
	RateLimitingConfig
}

type LoadBalancerConfig struct {
	StrategyName string `env:"BALANCER_STRATEGY_NAME,required"` // round_robin or least_connections
}

type HealthCheckConfig struct {
	Endpoint     string        `env:"HEALTH_CHECK_ENDPOINT,required"`
	BaseInterval time.Duration `env:"HEALTH_CHECK_BASE_INTERVAL,required"`
	MaxInterval  time.Duration `env:"HEALTH_CHECK_MAX_INTERVAL,required"`
}

type RateLimitingConfig struct {
	StrategyName string `env:"RATE_LIMITING_STRATEGY_NAME,required"` // token-bucket
}

// MustLoad loads configuration from environment variables.
// Throw a panic if the config doesn't exist or if there is an error reading the config.
func MustLoad() *Config {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("Failed to load config: %s", err))
	}

	return &cfg
}
