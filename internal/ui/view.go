package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	paneStyle          = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	activePaneStyle    = paneStyle.BorderForeground(lipgloss.Color("86"))
	previewStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1)
	activePreviewStyle = previewStyle.BorderForeground(lipgloss.Color("86"))
	statusStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func (m Model) View() string {
	if m.loading {
		return m.loadingView()
	}

	helpView := m.help.View(m)
	searchLine := ""
	if m.searching {
		searchLine = m.searchInput.View()
	} else if m.searchQuery != "" {
		searchLine = statusStyle.Render("Search: " + m.searchQuery)
	}

	footerHeight := lipgloss.Height(helpView) + 1
	if searchLine != "" {
		footerHeight++
	}
	bodyHeight := max(3, m.height-footerHeight)

	sourcePane := paneStyle
	articlePane := paneStyle
	previewPaneStyle := previewStyle

	switch m.activePane {
	case sourcesPane:
		sourcePane = activePaneStyle
	case articlesPane:
		articlePane = activePaneStyle
	case previewPane:
		previewPaneStyle = activePreviewStyle
	}

	sourceOuterWidth := m.sources.Width() + sourcePane.GetHorizontalFrameSize()
	articleOuterWidth := m.articles.Width() + articlePane.GetHorizontalFrameSize()
	previewInnerWidth := max(20, m.width-sourceOuterWidth-articleOuterWidth-previewPaneStyle.GetHorizontalFrameSize())

	sourcesView := renderBoundedPane(sourcePane, m.sources.View(), m.sources.Width(), bodyHeight)
	articlesView := renderBoundedPane(articlePane, m.articles.View(), m.articles.Width(), bodyHeight)
	previewView := renderBoundedPane(previewPaneStyle, m.previewText, previewInnerWidth, bodyHeight)

	content := lipgloss.JoinHorizontal(lipgloss.Top, sourcesView, articlesView, previewView)
	status := statusStyle.Render(m.statusText)

	if searchLine == "" {
		return fmt.Sprintf("%s\n%s\n%s", content, status, helpView)
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s", content, searchLine, status, helpView)
}

func (m Model) loadingView() string {
	spinnerText := m.spinner.View()
	msg := "Loading news..."

	line := spinnerText + " " + msg

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(line)
}

func renderBoundedPane(outer lipgloss.Style, raw string, innerWidth, totalHeight int) string {
	innerHeight := max(1, totalHeight-outer.GetVerticalFrameSize())
	content := lipgloss.NewStyle().
		Width(innerWidth).
		Height(innerHeight).
		MaxHeight(innerHeight).
		Render(raw)
	return outer.Render(content)
}
