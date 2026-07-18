package config

// Action identifies a bindable game action. Keys are stable strings stored in
// settings.json so bindings persist across versions.
type Action string

const (
	ActionMoveLeft   Action = "moveLeft"
	ActionMoveRight  Action = "moveRight"
	ActionSoftDrop   Action = "softDrop"
	ActionHardDrop   Action = "hardDrop"
	ActionRotateCW   Action = "rotateCW"
	ActionRotateCCW  Action = "rotateCCW"
	ActionRotate180  Action = "rotate180"
	ActionHold       Action = "hold"
	ActionPause      Action = "pause"
	ActionRestart    Action = "restart"
	ActionQuit       Action = "quit"
)

// DefaultKeyBindings maps each action to its default key (Bubble Tea key
// syntax: "left", "right", "up", "down", "space", "enter", or a single rune).
var DefaultKeyBindings = map[Action]string{
	ActionMoveLeft:   "left",
	ActionMoveRight:  "right",
	ActionSoftDrop:   "down",
	ActionHardDrop:   "space",
	ActionRotateCW:   "up",
	ActionRotateCCW:  "z",
	ActionRotate180:  "a",
	ActionHold:       "c",
	ActionPause:      "p",
	ActionRestart:    "r",
	ActionQuit:       "q",
}

// DefaultSettings returns the out-of-the-box configuration.
func DefaultSettings() Settings {
	return Settings{
		StartingLevel:    1,
		GhostEnabled:     true,
		HoldEnabled:      true,
		Rotate180Enabled: true,
		Theme:            "classic",
		FPSLimit:         60,
		KeyBindings:      cloneBindings(DefaultKeyBindings),
	}
}

func cloneBindings(m map[Action]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[string(k)] = v
	}
	return out
}
