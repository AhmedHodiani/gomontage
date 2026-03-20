package project

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the gomontage.yaml project configuration file.
type Config struct {
	// Name is the project name.
	Name string `yaml:"name"`

	// Version is the project version.
	Version string `yaml:"version"`

	// Resolution defines the output video dimensions.
	Resolution ResolutionConfig `yaml:"resolution"`

	// FPS is the output frame rate.
	FPS float64 `yaml:"fps"`

	// Output configures the output directory and format.
	Output OutputConfig `yaml:"output"`

	// Temp configures the temporary files directory.
	Temp TempConfig `yaml:"temp"`
}

// ResolutionConfig holds video resolution settings.
type ResolutionConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// OutputConfig holds output settings.
type OutputConfig struct {
	Directory string `yaml:"directory"`
	Format    string `yaml:"format"`
}

// TempConfig holds temporary file settings.
type TempConfig struct {
	Directory     string `yaml:"directory"`
	CleanOnExport bool   `yaml:"clean_on_export"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(name string) *Config {
	return &Config{
		Name:    name,
		Version: "1.0",
		Resolution: ResolutionConfig{
			Width:  1920,
			Height: 1080,
		},
		FPS: 30,
		Output: OutputConfig{
			Directory: "./output",
			Format:    "mp4",
		},
		Temp: TempConfig{
			Directory:     "./temp",
			CleanOnExport: true,
		},
	}
}

// LoadConfig reads a gomontage.yaml file and returns the parsed Config.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file %q: %w", path, err)
	}

	return &cfg, nil
}

// Save writes the Config to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write config file %q: %w", path, err)
	}

	return nil
}
