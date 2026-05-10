package app

import (
	"context"
	"fmt"

	"newsy/internal/domain"
	"newsy/internal/logging"
	"newsy/internal/search"
	"newsy/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) HandleAction(action ui.ActionMsg) tea.Cmd {
	switch action.Type {
	case ui.ActionRefreshCurrent:
		return a.runRefreshCurrent()
	case ui.ActionRefreshAll:
		return a.runRefreshAll()
	case ui.ActionToggleRead:
		return a.runToggleRead()
	case ui.ActionToggleStar:
		return a.runToggleStar()
	case ui.ActionOpen:
		return a.runOpen()
	case ui.ActionSearch:
		return a.runSearch(action.Query)
	case ui.ActionFilter:
		return a.runFilter()
	case ui.ActionClearSearch:
		return a.runClearSearch()
	case ui.ActionToggleUnread:
		return a.runToggleUnreadFilter()
	case ui.ActionToggleStarred:
		return a.runToggleStarredFilter()
	default:
		return nil
	}
}

func (a *App) runRefreshCurrent() tea.Cmd {
	spec, ok := a.model.SelectedSource()
	if !ok {
		return func() tea.Msg { return ui.ResultMsg{Status: "No source selected"} }
	}

	return func() tea.Msg {
		logging.Infof("manual refresh current source=%s", spec.Key)
		if err := a.service.RefreshSource(context.Background(), spec); err != nil {
			return ui.ResultMsg{Status: "Refresh failed", Err: err}
		}
		if err := a.reloadArticles(); err != nil {
			return ui.ResultMsg{Status: "Reload failed", Err: err}
		}
		return ui.ResultMsg{Status: fmt.Sprintf("Refreshed %s", spec.Name)}
	}
}

func (a *App) runRefreshAll() tea.Cmd {
	return func() tea.Msg {
		logging.Infof("manual refresh all begin sources=%d", len(a.specs))
		if err := a.service.RefreshAll(context.Background(), a.specs); err != nil {
			return ui.ResultMsg{Status: "Refresh all failed", Err: err}
		}
		if err := a.reloadArticles(); err != nil {
			return ui.ResultMsg{Status: "Reload failed", Err: err}
		}
		return ui.ResultMsg{Status: "All sources refreshed"}
	}
}

func (a *App) runToggleRead() tea.Cmd {
	article, ok := a.model.SelectedArticle()
	if !ok {
		return func() tea.Msg { return ui.ResultMsg{Status: "No article selected"} }
	}

	return func() tea.Msg {
		logging.Infof("toggle read article_id=%d new_value=%t", article.ID, !article.IsRead)
		if err := a.articles.SetRead(context.Background(), article.ID, !article.IsRead); err != nil {
			return ui.ResultMsg{Status: "Toggle read failed", Err: err}
		}
		if err := a.reloadArticles(); err != nil {
			return ui.ResultMsg{Status: "Reload failed", Err: err}
		}
		return ui.ResultMsg{Status: "Read state updated"}
	}
}

func (a *App) runToggleStar() tea.Cmd {
	article, ok := a.model.SelectedArticle()
	if !ok {
		return func() tea.Msg { return ui.ResultMsg{Status: "No article selected"} }
	}

	return func() tea.Msg {
		logging.Infof("toggle star article_id=%d new_value=%t", article.ID, !article.IsStarred)
		if err := a.articles.SetStarred(context.Background(), article.ID, !article.IsStarred); err != nil {
			return ui.ResultMsg{Status: "Toggle star failed", Err: err}
		}
		if err := a.reloadArticles(); err != nil {
			return ui.ResultMsg{Status: "Reload failed", Err: err}
		}
		return ui.ResultMsg{Status: "Star state updated"}
	}
}

func (a *App) runOpen() tea.Cmd {
	article, ok := a.model.SelectedArticle()
	if !ok {
		return func() tea.Msg { return ui.ResultMsg{Status: "No article selected"} }
	}

	return func() tea.Msg {
		logging.Infof("open article article_id=%d link=%s", article.ID, article.Link)
		if err := a.OpenArticle(article); err != nil {
			return ui.ResultMsg{Status: "Open failed", Err: err}
		}
		return ui.ResultMsg{Status: "Opened article in browser"}
	}
}

func (a *App) runSearch(query string) tea.Cmd {
	return func() tea.Msg {
		a.model.SetSearchQuery(query)
		a.applyFilters()
		if query == "" {
			return ui.ResultMsg{Status: "Search cleared"}
		}
		return ui.ResultMsg{Status: fmt.Sprintf("Search applied: %q", query)}
	}
}

func (a *App) runClearSearch() tea.Cmd {
	return func() tea.Msg {
		a.model.SetSearchQuery("")
		a.applyFilters()
		return ui.ResultMsg{Status: "Search cleared"}
	}
}

func (a *App) runToggleUnreadFilter() tea.Cmd {
	return func() tea.Msg {
		enabled := a.model.ToggleUnreadFilter()
		a.applyFilters()
		if enabled {
			return ui.ResultMsg{Status: "Unread filter enabled"}
		}
		return ui.ResultMsg{Status: "Unread filter disabled"}
	}
}

func (a *App) runToggleStarredFilter() tea.Cmd {
	return func() tea.Msg {
		enabled := a.model.ToggleStarredFilter()
		a.applyFilters()
		if enabled {
			return ui.ResultMsg{Status: "Starred filter enabled"}
		}
		return ui.ResultMsg{Status: "Starred filter disabled"}
	}
}

func (a *App) runFilter() tea.Cmd {
	return func() tea.Msg {
		a.applyFilters()
		name := a.model.SelectedSourceName()
		if name == "" {
			return ui.ResultMsg{Status: "Source filter applied"}
		}
		return ui.ResultMsg{Status: fmt.Sprintf("Showing %s", name)}
	}
}

func (a *App) reloadArticles() error {
	articles, err := a.articles.ListBySource(context.Background(), "")
	if err != nil {
		return err
	}
	a.allArticles = articles
	a.applyFilters()
	return nil
}

func (a *App) applyFilters() {
	a.model.SetSources(a.specs, countArticlesBySource(a.allArticles))
	sourceKey := a.model.SelectedSourceKey()
	filtered := filterBySource(a.allArticles, sourceKey)
	filtered = filterByState(filtered, a.model.OnlyUnread(), a.model.OnlyStarred())
	filtered = search.FilterByTitle(filtered, a.model.SearchQuery())
	a.model.SetArticles(filtered)
}

func filterBySource(articles []domain.Article, sourceKey string) []domain.Article {
	if sourceKey == "" {
		return articles
	}
	filtered := make([]domain.Article, 0, len(articles))
	for _, article := range articles {
		if article.SourceKey == sourceKey {
			filtered = append(filtered, article)
		}
	}
	return filtered
}

func filterByState(articles []domain.Article, onlyUnread, onlyStarred bool) []domain.Article {
	if !onlyUnread && !onlyStarred {
		return articles
	}
	filtered := make([]domain.Article, 0, len(articles))
	for _, article := range articles {
		if onlyUnread && article.IsRead {
			continue
		}
		if onlyStarred && !article.IsStarred {
			continue
		}
		filtered = append(filtered, article)
	}
	return filtered
}
