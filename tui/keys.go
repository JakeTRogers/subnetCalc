package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines the keyboard bindings for the TUI.
// It implements the help.KeyMap interface for displaying key hints.
type keyMap struct {
	Up       key.Binding // Move cursor up
	Down     key.Binding // Move cursor down
	Left     key.Binding // Scroll split columns left
	Right    key.Binding // Scroll split columns right
	PageUp   key.Binding // Page up through rows
	PageDown key.Binding // Page down through rows
	Split    key.Binding // Split selected subnet
	Join     key.Binding // Join selected subnet with sibling
	Undo     key.Binding // Undo last split/join
	Redo     key.Binding // Redo last undone operation
	Export   key.Binding // Export tree as JSON
	Copy     key.Binding // Copy JSON to clipboard
	Quit     key.Binding // Exit the TUI
	Help     key.Binding // Toggle help display
}

// ShortHelp returns key bindings for the short help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.PageUp, k.PageDown, k.Split, k.Join, k.Undo, k.Redo, k.Export, k.Help, k.Quit}
}

// FullHelp returns key bindings for the full help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Left, k.Right},
		{k.Split, k.Join, k.Undo, k.Redo},
		{k.Export, k.Copy},
		{k.Help, k.Quit},
	}
}

// defaultKeys provides the default key bindings.
var defaultKeys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "scroll left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "scroll right"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),
	Split: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "split"),
	),
	Join: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "join"),
	),
	Undo: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "undo"),
	),
	Redo: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "redo"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export JSON"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy to clipboard"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
