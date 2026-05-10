package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if !m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		return m, nil
	case ResultMsg:
		if msg.Err != nil {
			m.statusText = msg.Status + ": " + msg.Err.Error()
		} else {
			m.statusText = msg.Status
		}
		return m, nil
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "esc":
				m.StopSearch()
				m.statusText = "Search cancelled"
				m.resize()
				return m, nil
			case "enter":
				query := m.SearchQuery()
				m.SetSearchQuery(query)
				m.StopSearch()
				m.resize()
				return m, EmitActionWithQuery(ActionSearch, query)
			}

			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		switch {
		case keyMatches(msg, m.keymap.Quit):
			return m, tea.Quit
		case keyMatches(msg, m.keymap.NextPane):
			m.activePane = (m.activePane + 1) % 3
			return m, nil
		case keyMatches(msg, m.keymap.PrevPane):
			if m.activePane == 0 {
				m.activePane = 2
			} else {
				m.activePane--
			}
			return m, nil
		case keyMatches(msg, m.keymap.RefreshAll):
			m.statusText = "Refreshing all sources..."
			return m, EmitAction(ActionRefreshAll)
		case keyMatches(msg, m.keymap.Refresh):
			m.statusText = "Refreshing current source..."
			return m, EmitAction(ActionRefreshCurrent)
		case keyMatches(msg, m.keymap.ToggleRead):
			return m, EmitAction(ActionToggleRead)
		case keyMatches(msg, m.keymap.ToggleStar):
			return m, EmitAction(ActionToggleStar)
		case keyMatches(msg, m.keymap.Open):
			return m, EmitAction(ActionOpen)
		case keyMatches(msg, m.keymap.Search):
			m.StartSearch()
			m.statusText = "Enter search and press Enter"
			m.resize()
			return m, nil
		case keyMatches(msg, m.keymap.ClearSearch):
			if m.searchQuery == "" {
				return m, nil
			}
			return m, EmitAction(ActionClearSearch)
		case keyMatches(msg, m.keymap.ToggleUnread):
			return m, EmitAction(ActionToggleUnread)
		case keyMatches(msg, m.keymap.ToggleStarred):
			return m, EmitAction(ActionToggleStarred)
		}
	}

	var cmd tea.Cmd
	switch m.activePane {
	case sourcesPane:
		before := m.sources.Index()
		m.sources, cmd = m.sources.Update(msg)
		if m.sources.Index() != before {
			return m, EmitAction(ActionFilter)
		}
	case articlesPane:
		m.articles, cmd = m.articles.Update(msg)
		m.updatePreview()
	}

	return m, cmd
}

func (m *Model) resize() {
	if m.width == 0 || m.height == 0 {
		return
	}

	footerHeight := lipgloss.Height(m.help.View(*m)) + 1
	if m.searching || m.searchQuery != "" {
		footerHeight++
	}
	bodyHeight := max(3, m.height-footerHeight)

	sourceOuterWidth := max(20, m.width/4)
	articleOuterWidth := max(30, m.width/3)
	sourceInnerWidth := max(1, sourceOuterWidth-paneStyle.GetHorizontalFrameSize())
	articleInnerWidth := max(1, articleOuterWidth-paneStyle.GetHorizontalFrameSize())
	innerHeight := max(1, bodyHeight-paneStyle.GetVerticalFrameSize())

	m.sources.SetSize(sourceInnerWidth, innerHeight)
	m.articles.SetSize(articleInnerWidth, innerHeight)
}
