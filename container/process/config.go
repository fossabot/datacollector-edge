package process

const (
	DefaultProcessMetricsCaptureInterval = "-1"
)

type Config struct {
	ProcessMetricsCaptureInterval string `toml:"process-metrics-capture-interval"`
}

// NewConfig returns a new Config with default settings.
func NewConfig() Config {
	return Config{
		ProcessMetricsCaptureInterval: DefaultProcessMetricsCaptureInterval,
	}
}
