package internal

// Check is a configured check ready to be added to a runner.
type Check interface {
	AddtoRunner(*Runner)
}

// CheckConfig defines a Check.
type CheckConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Kind        string            `yaml:"kind"`
	Spec        map[string]string `yaml:"spec"`
}

// ReliabilityConfig defines an overall configuration for the Reliability Scanner.
type ReliabilityConfig struct {
	Checks []CheckConfig `yaml:"checks"`
}

// Config is used to read config back in at runtime.
type Config struct {
	Cfg ReliabilityConfig `env:"CONFIG"`
}
