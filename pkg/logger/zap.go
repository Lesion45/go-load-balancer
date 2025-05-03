package logger

import "go.uber.org/zap"

const (
	logDev  = "dev"
	logProd = "prod"
)

// NewZap returns a new instance of the zap.Logger.
//
// newLogger creates a zap.Logger based on the "ENV" variable.
// Uses zap.NewDevelopment() for "dev" and zap.NewProduction() for "prod".
// Defaults to zap.NewProduction() for invalid or missing "ENV" values.
func NewZap(logLevel string) *zap.Logger {
	var cfg zap.Config

	switch logLevel {
	case logDev:
		cfg = zap.NewDevelopmentConfig()
	case logProd:
		cfg = zap.NewProductionConfig()
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.DisableStacktrace = true

	cfg.DisableCaller = true

	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return log
}
