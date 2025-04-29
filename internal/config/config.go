package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3" // Using yaml.v3 for parsing
)

// DefaultsConfig holds default startup settings
type DefaultsConfig struct {
	Preset string `yaml:"preset"`
}

// GUIConfig holds GUI related settings
type GUIConfig struct {
	Theme            string `yaml:"theme"`
	PlotUpdateRateHz int    `yaml:"plotUpdateRateHz"`
}

// AnalyzerConfig holds settings for the frequency analyzer
type AnalyzerConfig struct {
	FFTSize      int     `yaml:"fftSize"`
	MinFrequency float64 `yaml:"minFrequency"`
	MaxFrequency float64 `yaml:"maxFrequency"`
}

// Config holds the overall application configuration
type Config struct {
	Defaults DefaultsConfig `yaml:"defaults"`
	GUI      GUIConfig      `yaml:"gui"`
	Analyzer AnalyzerConfig `yaml:"analyzer"`
}

var (
	currentConfig *Config
	configOnce    sync.Once
	configMutex   sync.RWMutex
)

// LoadConfig reads configuration from a YAML file
func LoadConfig(filePath string) (*Config, error) {
	var err error
	configOnce.Do(func() {
		cfg := &Config{ // Set default values here in case file loading fails
			Defaults: DefaultsConfig{Preset: "Default"},
			GUI:      GUIConfig{Theme: "dark", PlotUpdateRateHz: 30},
			Analyzer: AnalyzerConfig{FFTSize: 2048, MinFrequency: 20.0, MaxFrequency: 20000.0},
		}

		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			// Log error but use defaults
			err = readErr // Store the error to return it
			currentConfig = cfg
			return
		}

		yamlErr := yaml.Unmarshal(data, cfg)
		if yamlErr != nil {
			// Log error but use defaults
			err = yamlErr // Store the error to return it
			currentConfig = cfg
			return
		}
		currentConfig = cfg
	})

	configMutex.RLock()
	defer configMutex.RUnlock()
	// Return the potentially modified error from the Do func
	return currentConfig, err
}

// GetConfig returns the loaded configuration (thread-safe)
func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if currentConfig == nil {
		// Should not happen if LoadConfig was called, but return defaults as fallback
		return &Config{
			Defaults: DefaultsConfig{Preset: "Default"},
			GUI:      GUIConfig{Theme: "dark", PlotUpdateRateHz: 30},
			Analyzer: AnalyzerConfig{FFTSize: 2048, MinFrequency: 20.0, MaxFrequency: 20000.0},
		}
	}
	return currentConfig
}

// SaveConfig saves the current configuration back to the file (optional)
func SaveConfig(filePath string) error {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if currentConfig == nil {
		return nil // Nothing to save
	}

	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
