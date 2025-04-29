package presets

import (
	"encoding/json"
	"os"
	"sync"
)

// BandSettings defines the parameters for a single EQ band in a preset
type BandSettings struct {
	Type    string  `json:"type"` // e.g., "Peak", "LowShelf"
	Freq    float64 `json:"freq"`
	Q       float64 `json:"q"`
	GainDB  float64 `json:"gainDB"`
	Enabled bool    `json:"enabled"`
}

// Preset defines the structure of a single EQ preset
type Preset struct {
	MasterGain float32        `json:"masterGain"` // VST parameter value (0.0 to 1.0)
	Bands      []BandSettings `json:"bands"`
}

// Manager handles loading and accessing presets
type Manager struct {
	presets map[string]Preset
	mutex   sync.RWMutex
}

// NewManager creates a preset manager and loads presets from a file
func NewManager(filePath string) (*Manager, error) {
	m := &Manager{
		presets: make(map[string]Preset),
	}
	err := m.LoadPresets(filePath)
	if err != nil {
		// Log error but potentially continue with default/empty presets
		// return nil, fmt.Errorf("failed to load presets from %s: %w", filePath, err)
	}
	return m, nil
}

// LoadPresets reads presets from a JSON file
func (m *Manager) LoadPresets(filePath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &m.presets)
	if err != nil {
		return err
	}
	return nil
}

// GetPreset retrieves a preset by name
func (m *Manager) GetPreset(name string) (Preset, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	preset, ok := m.presets[name]
	return preset, ok
}

// ListPresets returns the names of all loaded presets
func (m *Manager) ListPresets() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	names := make([]string, 0, len(m.presets))
	for name := range m.presets {
		names = append(names, name)
	}
	return names
}

// AddPreset adds or updates a preset (useful for saving user presets)
func (m *Manager) AddPreset(name string, preset Preset) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.presets[name] = preset
	// TODO: Add functionality to save presets back to the file if needed
}
