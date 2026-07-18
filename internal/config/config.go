package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings is the persisted user configuration.
type Settings struct {
	StartingLevel    int            `json:"startingLevel"`
	GhostEnabled     bool           `json:"ghostEnabled"`
	HoldEnabled      bool           `json:"holdEnabled"`
	Rotate180Enabled bool           `json:"rotate180Enabled"`
	Theme            string         `json:"theme"`
	FPSLimit         int            `json:"fpsLimit"`
	KeyBindings      map[string]string `json:"keyBindings"`
}

// Filename is where settings are stored (next to the executable's working dir).
const Filename = "settings.json"

// Load reads settings.json, merging any missing keys with defaults so new
// options added in future versions still get sane values.
func Load(path string) (Settings, error) {
	if path == "" {
		path = Filename
	}
	def := DefaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return def, nil
		}
		return def, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return def, err
	}
	mergeDefaults(&s, def)
	return s, nil
}

// Save writes settings to settings.json, creating parent dirs as needed.
func Save(s Settings, path string) error {
	if path == "" {
		path = Filename
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func mergeDefaults(s *Settings, def Settings) {
	if s.StartingLevel < 1 {
		s.StartingLevel = def.StartingLevel
	}
	if s.Theme == "" {
		s.Theme = def.Theme
	}
	if s.FPSLimit < 1 {
		s.FPSLimit = def.FPSLimit
	}
	if s.KeyBindings == nil {
		s.KeyBindings = def.KeyBindings
	} else {
		for k, v := range def.KeyBindings {
			if _, ok := s.KeyBindings[k]; !ok {
				s.KeyBindings[k] = v
			}
		}
	}
}

// Binding returns the configured key for an action, falling back to default.
func (s Settings) Binding(a Action) string {
	if k, ok := s.KeyBindings[string(a)]; ok && k != "" {
		return k
	}
	return DefaultKeyBindings[a]
}
