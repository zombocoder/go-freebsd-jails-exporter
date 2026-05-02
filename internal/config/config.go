package config

import "time"

type Config struct {
	Web       WebConfig
	Log       LogConfig
	Collector CollectorConfig
}

type WebConfig struct {
	ListenAddress string
	TelemetryPath string
}

type LogConfig struct {
	Level  string
	Format string
}

type CollectorConfig struct {
	JailsEnabled     bool
	ProcessesEnabled bool
	RACCTEnabled     bool
	Timeout          time.Duration
	RACCTTimeout     time.Duration
}

func Default() *Config {
	return &Config{
		Web: WebConfig{
			ListenAddress: ":9756",
			TelemetryPath: "/metrics",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "logfmt",
		},
		Collector: CollectorConfig{
			JailsEnabled:     true,
			ProcessesEnabled: true,
			RACCTEnabled:     true,
			Timeout:          5 * time.Second,
			RACCTTimeout:     2 * time.Second,
		},
	}
}
