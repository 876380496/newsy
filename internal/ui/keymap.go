package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	PrevPane      key.Binding
	NextPane      key.Binding
	Refresh       key.Binding
	RefreshAll    key.Binding
	ToggleRead    key.Binding
	ToggleStar    key.Binding
	Search        key.Binding
	ClearSearch   key.Binding
	ToggleUnread  key.Binding
	ToggleStarred key.Binding
	Open          key.Binding
	Quit          key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/↓", "down"),
		),
		PrevPane: key.NewBinding(
			key.WithKeys("shift+tab", "h", "left"),
			key.WithHelp("h/shift+tab/←", "prev pane"),
		),
		NextPane: key.NewBinding(
			key.WithKeys("tab", "l", "right"),
			key.WithHelp("l/tab/→", "next pane"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh source"),
		),
		RefreshAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh all"),
		),
		ToggleRead: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "toggle read"),
		),
		ToggleStar: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle star"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear search"),
		),
		ToggleUnread: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "toggle unread"),
		),
		ToggleStarred: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle starred"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
