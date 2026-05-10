package app

import (
	"context"

	"newsy/internal/browser"
	"newsy/internal/config"
	"newsy/internal/domain"
	"newsy/internal/logging"
	"newsy/internal/providers"
	"newsy/internal/runtimepaths"
	"newsy/internal/source"
	"newsy/internal/store"
	"newsy/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type articleRepo interface {
	ListBySource(ctx context.Context, sourceKey string) ([]domain.Article, error)
	SetRead(ctx context.Context, id int64, isRead bool) error
	SetStarred(ctx context.Context, id int64, isStarred bool) error
}

type App struct {
	program      *tea.Program
	store        *store.SQLiteStore
	instanceLock *instanceLock
	paths        runtimepaths.Paths
	specs        []source.Spec
	service      *source.Service
	articles     articleRepo
	allArticles  []domain.Article
	model        ui.Model
}

func New(paths runtimepaths.Paths) (*App, error) {
	logging.Infof("app.New start")
	appLock, err := acquireInstanceLock(paths.LockFile)
	if err != nil {
		logging.Errorf("instance lock acquire failed path=%s err=%v", paths.LockFile, err)
		return nil, err
	}

	cfg, err := config.Load(paths.ConfigFile)
	if err != nil {
		logging.Errorf("config load failed path=%s err=%v", paths.ConfigFile, err)
		_ = appLock.Close()
		return nil, err
	}
	logging.Infof("config loaded path=%s sources=%d", paths.ConfigFile, len(cfg.Sources))

	specs := make([]source.Spec, 0, len(cfg.Sources))
	for _, src := range cfg.Sources {
		specs = append(specs, source.SpecFromConfig(src))
	}

	db, err := store.Open(paths.DBFile)
	if err != nil {
		logging.Errorf("sqlite open failed path=%s err=%v", paths.DBFile, err)
		_ = appLock.Close()
		return nil, err
	}
	logging.Infof("sqlite opened path=%s", paths.DBFile)
	articleRepo := store.NewArticleRepository(db)
	stateRepo := store.NewStateRepository(db)
	registry := source.NewRegistry()
	if err := providers.RegisterBuiltins(registry, paths.PluginDir); err != nil {
		logging.Errorf("register builtins failed plugin_dir=%s err=%v", paths.PluginDir, err)
		_ = db.Close()
		_ = appLock.Close()
		return nil, err
	}
	logging.Infof("providers registered plugin_dir=%s", paths.PluginDir)
	if err := source.ValidateSpecs(registry, specs); err != nil {
		logging.Errorf("validate specs failed: %v", err)
		_ = db.Close()
		_ = appLock.Close()
		return nil, err
	}
	logging.Infof("source specs validated")
	service := source.NewService(registry, articleRepo, stateRepo)

	model := ui.NewModel()

	app := &App{
		store:        db,
		instanceLock: appLock,
		paths:        paths,
		specs:        specs,
		service:      service,
		articles:     articleRepo,
		allArticles:  nil,
		model:        model,
	}

	app.program = tea.NewProgram(app, tea.WithAltScreen())
	logging.Infof("app.New complete")
	return app, nil
}

func (a *App) startupLoad() tea.Cmd {
	return func() tea.Msg {
		logging.Infof("startup cache load begin")
		articles, err := a.articles.ListBySource(context.Background(), "")
		if err != nil {
			logging.Errorf("startup cache load failed: %v", err)
			return ui.InitialDataMsg{Err: err}
		}
		logging.Infof("startup cache load complete articles=%d", len(articles))
		return ui.InitialDataMsg{Articles: articles}
	}
}

func (a *App) startupRefresh() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(a.specs))
	for _, spec := range a.specs {
		if !spec.Enabled {
			continue
		}
		spec := spec
		cmds = append(cmds, func() tea.Msg {
			logging.Infof("startup refresh begin source=%s provider=%s", spec.Key, spec.ProviderType)
			err := a.service.RefreshSource(context.Background(), spec)
			return ui.SourceRefreshMsg{SourceKey: spec.Key, Name: spec.Name, Err: err}
		})
	}
	return tea.Batch(cmds...)
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.model.Init(),
		a.startupLoad(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ui.InitialDataMsg:
		if msg.Err != nil {
			logging.Errorf("initial data message error: %v", msg.Err)
			a.model.SetStatus("Initial load: " + msg.Err.Error())
		} else {
			logging.Infof("initial data received articles=%d", len(msg.Articles))
			a.allArticles = msg.Articles
		}
		for _, spec := range a.specs {
			if spec.Enabled {
				a.model.SetSourceState(spec.Key, "loading")
			}
		}
		a.model.SetSources(a.specs, countArticlesBySource(a.allArticles))
		a.applyFilters()
		a.model.SetLoading(false)
		return a, a.startupRefresh()
	case ui.SourceRefreshMsg:
		state := ""
		if msg.Err != nil {
			logging.Errorf("source refresh failed source=%s name=%s err=%v", msg.SourceKey, msg.Name, msg.Err)
			state = msg.Err.Error()
		} else {
			logging.Infof("source refresh succeeded source=%s name=%s", msg.SourceKey, msg.Name)
		}
		a.model.SetSourceState(msg.SourceKey, state)
		if msg.Err == nil {
			_ = a.reloadArticles()
		}
		a.model.SetSources(a.specs, countArticlesBySource(a.allArticles))
		if msg.Err == nil {
			a.applyFilters()
		}
		return a, nil
	case ui.ActionMsg:
		return a, a.HandleAction(msg)
	default:
		nextModel, cmd := a.model.Update(msg)
		a.model = nextModel.(ui.Model)
		return a, cmd
	}
}

func (a *App) View() string {
	return a.model.View()
}

func (a *App) Run() error {
	defer a.store.Close()
	defer func() {
		if a.instanceLock != nil {
			_ = a.instanceLock.Close()
		}
	}()
	_, err := a.program.Run()
	return err
}

func (a *App) OpenArticle(article domain.Article) error {
	return browser.Open(article.Link)
}

func countArticlesBySource(articles []domain.Article) map[string]int {
	counts := make(map[string]int, len(articles))
	for _, article := range articles {
		counts[article.SourceKey]++
	}
	return counts
}
