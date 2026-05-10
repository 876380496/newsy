package ui

import (
	"newsy/internal/domain"

	tea "github.com/charmbracelet/bubbletea"
)

type ActionType string

const (
	ActionRefreshCurrent ActionType = "refresh_current"
	ActionRefreshAll     ActionType = "refresh_all"
	ActionToggleRead     ActionType = "toggle_read"
	ActionToggleStar     ActionType = "toggle_star"
	ActionOpen           ActionType = "open"
	ActionSearch         ActionType = "search"
	ActionFilter         ActionType = "filter"
	ActionClearSearch    ActionType = "clear_search"
	ActionToggleUnread   ActionType = "toggle_unread"
	ActionToggleStarred  ActionType = "toggle_starred"
)

// InitialDataMsg carries the result of the initial async data fetch at startup.
type InitialDataMsg struct {
	Articles []domain.Article
	Err      error
}

// SourceRefreshMsg is sent when a single source finishes its background refresh.
type SourceRefreshMsg struct {
	SourceKey string
	Name      string
	Err       error
}

type ActionMsg struct {
	Type  ActionType
	Query string
}

type ResultMsg struct {
	Status string
	Err    error
}

func EmitAction(actionType ActionType) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{Type: actionType}
	}
}

func EmitActionWithQuery(actionType ActionType, query string) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{Type: actionType, Query: query}
	}
}
