package input

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mahir/tetris/internal/config"
)

// Bindings resolves configured key bindings to game actions. Matching is done
// via explicit msg.Type checks so quirks like KeySpace.String() == " " do not
// break binding resolution.
type Bindings struct {
	byToken map[string]config.Action
}

// NewBindings builds a resolver from settings.
func NewBindings(s config.Settings) *Bindings {
	m := make(map[string]config.Action, len(s.KeyBindings))
	for act, k := range s.KeyBindings {
		m[k] = config.Action(act)
	}
	return &Bindings{byToken: m}
}

// token returns a normalized key token for a key message.
func token(msg tea.KeyMsg) string {
	switch msg.Type {
	case tea.KeySpace:
		return "space"
	case tea.KeyEnter:
		return "enter"
	case tea.KeyEsc:
		return "esc"
	case tea.KeyUp:
		return "up"
	case tea.KeyDown:
		return "down"
	case tea.KeyLeft:
		return "left"
	case tea.KeyRight:
		return "right"
	case tea.KeyRunes:
		return string(msg.Runes)
	default:
		return msg.String()
	}
}

// Match returns the action bound to a key message, or "" if unbound.
func (b *Bindings) Match(msg tea.KeyMsg) config.Action {
	if a, ok := b.byToken[token(msg)]; ok {
		return a
	}
	return ""
}

// IsAction reports whether the key message is bound to the given action.
func (b *Bindings) IsAction(msg tea.KeyMsg, a config.Action) bool {
	return b.Match(msg) == a
}
