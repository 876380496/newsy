package ui

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"newsy/internal/domain"
	"newsy/internal/source"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pane int

const (
	sourcesPane pane = iota
	articlesPane
	previewPane
)

var (
	previewBlockTagPattern = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|h[1-6]|article|section|blockquote)[^>]*>`)
	previewTagPattern      = regexp.MustCompile(`(?s)<[^>]+>`)
)

type Model struct {
	width          int
	height         int
	activePane     pane
	sources        list.Model
	articles       list.Model
	help           help.Model
	keymap         KeyMap
	previewText    string
	statusText     string
	searchInput    textinput.Model
	searching      bool
	searchQuery    string
	onlyUnread     bool
	onlyStarred    bool
	sourceSpecs    []source.Spec
	articleEntries []domain.Article
	loading        bool
	spinner        spinner.Model
	sourceStates   map[string]string
}

type listItem struct {
	title string
	desc  string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }

func NewModel() Model {
	keymap := DefaultKeyMap()

	sourceDelegate := list.NewDefaultDelegate()
	sourceDelegate.SetSpacing(0)
	sourceDelegate.SetHeight(1)
	sourceDelegate.ShowDescription = false
	sources := list.New([]list.Item{
		listItem{title: "No sources configured", desc: "Add sources in config to get started"},
	}, sourceDelegate, 0, 0)
	sources.SetShowHelp(false)
	sources.SetShowPagination(false)
	sources.SetShowTitle(false)
	sources.SetShowStatusBar(false)

	articleDelegate := list.NewDefaultDelegate()
	articleDelegate.SetSpacing(0)
	articleDelegate.SetHeight(1)
	articleDelegate.ShowDescription = false
	articles := list.New([]list.Item{
		listItem{title: "No articles yet", desc: "Refresh after configuring a source"},
	}, articleDelegate, 0, 0)
	articles.SetShowHelp(false)
	articles.SetShowPagination(false)
	articles.SetShowTitle(false)
	articles.SetShowStatusBar(false)

	searchInput := textinput.New()
	searchInput.Placeholder = "Search article titles"
	searchInput.Prompt = "/ "

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	s.Spinner = spinner.Dot

	return Model{
		activePane:     sourcesPane,
		sources:        sources,
		articles:       articles,
		help:           help.New(),
		keymap:         keymap,
		previewText:    "Preview will appear here.",
		statusText:     "Loading...",
		searchInput:    searchInput,
		articleEntries: nil,
		loading:        true,
		spinner:        s,
		sourceStates:   make(map[string]string),
	}
}

func (m Model) Init() tea.Cmd {
	if m.loading {
		return m.spinner.Tick
	}
	return nil
}

func (m *Model) SetLoading(v bool) {
	m.loading = v
}

func (m *Model) SetSourceState(key, state string) {
	if m.sourceStates == nil {
		m.sourceStates = make(map[string]string)
	}
	m.sourceStates[key] = state
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keymap.Up, m.keymap.Down, m.keymap.NextPane, m.keymap.Refresh, m.keymap.Quit}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.keymap.Up, m.keymap.Down, m.keymap.PrevPane, m.keymap.NextPane}, {m.keymap.Refresh, m.keymap.RefreshAll, m.keymap.Search, m.keymap.ClearSearch}, {m.keymap.ToggleUnread, m.keymap.ToggleStarred, m.keymap.ToggleRead, m.keymap.ToggleStar}, {m.keymap.Open, m.keymap.Quit}}
}

func (m *Model) SetSources(specs []source.Spec, counts map[string]int) {
	selectedKey := m.SelectedSourceKey()
	m.sourceSpecs = specs
	items := make([]list.Item, 0, len(specs))
	selectedIndex := 0
	for i, spec := range specs {
		status := spec.ProviderType
		if !spec.Enabled {
			status += " · disabled"
		}
		title := renderSourceTitle(spec, counts[spec.Key], m.sourceStates[spec.Key])
		items = append(items, listItem{title: title, desc: status})
		if spec.Key == selectedKey {
			selectedIndex = i
		}
	}
	if len(items) == 0 {
		items = []list.Item{listItem{title: "No sources configured", desc: "Add sources in config to get started"}}
		selectedIndex = 0
	}
	m.sources.SetItems(items)
	m.sources.Select(selectedIndex)
}

func (m *Model) SetArticles(articles []domain.Article) {
	m.articleEntries = articles
	items := make([]list.Item, 0, len(articles))
	for _, article := range articles {
		prefix := ""
		if article.IsStarred {
			prefix += "★ "
		}
		title := strings.TrimSpace(article.Title)
		if title == "" {
			title = "(untitled)"
		}
		items = append(items, listItem{title: formatArticleTime(article) + prefix + title, desc: article.SourceKey})
	}
	if len(items) == 0 {
		items = []list.Item{listItem{title: "No articles found", desc: "Try another source or search"}}
	}
	m.articles.SetItems(items)
	m.articles.Select(0)
	m.updatePreview()
}

var beijingLocation = time.FixedZone("CST", 8*3600)

func formatArticleTime(article domain.Article) string {
	if article.PublishedAt.IsZero() {
		return ""
	}
	return "  " + article.PublishedAt.In(beijingLocation).Format("01-02 15:04")
}

func renderSourceTitle(spec source.Spec, count int, state string) string {
	if state == "loading" {
		return fmt.Sprintf("%s(loading...)", spec.Name)
	}
	if state != "" {
		return fmt.Sprintf("%s(error)", spec.Name)
	}
	return fmt.Sprintf("%s(%d篇)", spec.Name, count)
}

func (m *Model) SetStatus(status string) {
	if strings.TrimSpace(status) == "" {
		m.statusText = "Ready"
		return
	}
	m.statusText = status
}

func (m *Model) StartSearch() {
	m.searching = true
	m.searchInput.SetValue(m.searchQuery)
	m.searchInput.Focus()
}

func (m *Model) StopSearch() {
	m.searching = false
	m.searchInput.Blur()
}

func (m *Model) SearchQuery() string {
	return strings.TrimSpace(m.searchInput.Value())
}

func (m *Model) SelectedSourceKey() string {
	spec, ok := m.SelectedSource()
	if !ok {
		return ""
	}
	return spec.Key
}

func (m *Model) SelectedSourceName() string {
	spec, ok := m.SelectedSource()
	if !ok {
		return ""
	}
	return spec.Name
}

func (m *Model) SetSearchQuery(query string) {
	m.searchQuery = strings.TrimSpace(query)
	m.searchInput.SetValue(m.searchQuery)
}

func (m *Model) ToggleUnreadFilter() bool {
	m.onlyUnread = !m.onlyUnread
	return m.onlyUnread
}

func (m *Model) ToggleStarredFilter() bool {
	m.onlyStarred = !m.onlyStarred
	return m.onlyStarred
}

func (m Model) OnlyUnread() bool {
	return m.onlyUnread
}

func (m Model) OnlyStarred() bool {
	return m.onlyStarred
}

func (m Model) FilterSummary() string {
	parts := make([]string, 0, 3)
	if m.searchQuery != "" {
		parts = append(parts, "search="+m.searchQuery)
	}
	if m.onlyUnread {
		parts = append(parts, "unread")
	}
	if m.onlyStarred {
		parts = append(parts, "starred")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ")
}

func (m *Model) updatePreview() {
	index := m.articles.Index()
	if index < 0 || index >= len(m.articleEntries) {
		m.previewText = "Preview will appear here."
		return
	}

	article := m.articleEntries[index]
	body := article.Content
	if strings.TrimSpace(body) == "" {
		body = article.Summary
	}
	body = cleanPreviewText(body)
	if body == "" {
		body = "No preview content available."
	}

	title := strings.TrimSpace(article.Title)
	if title == "" {
		title = "(untitled)"
	}
	m.previewText = fmt.Sprintf("%s\n%s\n\n%s", title, article.Link, body)
}

func cleanPreviewText(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}

	text = previewBlockTagPattern.ReplaceAllString(text, "\n")
	text = previewTagPattern.ReplaceAllString(text, "")
	text = html.UnescapeString(text)

	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

func (m Model) SelectedSource() (source.Spec, bool) {
	index := m.sources.Index()
	if index < 0 || index >= len(m.sourceSpecs) {
		return source.Spec{}, false
	}
	return m.sourceSpecs[index], true
}

func (m Model) SelectedArticle() (domain.Article, bool) {
	index := m.articles.Index()
	if index < 0 || index >= len(m.articleEntries) {
		return domain.Article{}, false
	}
	return m.articleEntries[index], true
}
