package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"newsy/internal/domain"
	"newsy/internal/logging"
	"newsy/internal/source"
)

const (
	defaultFetchTimeout    = 30 * time.Second
	defaultValidateTimeout = 10 * time.Second
)

// ExecProvider implements source.Provider by delegating to external executables.
// Type() always returns "plugin"; which executable to run is specified per-source
// via config["plugin"] (name in plugins dir) or config["plugin_path"] (explicit path).
type ExecProvider struct {
	pluginsDir string
	timeout    time.Duration
}

// NewExecProvider creates an ExecProvider that discovers plugins in the given directory.
func NewExecProvider(pluginsDir string) *ExecProvider {
	return &ExecProvider{
		pluginsDir: pluginsDir,
		timeout:    defaultFetchTimeout,
	}
}

// NewExecProviderWithTimeout creates an ExecProvider with a custom fetch timeout.
func NewExecProviderWithTimeout(pluginsDir string, timeout time.Duration) *ExecProvider {
	return &ExecProvider{
		pluginsDir: pluginsDir,
		timeout:    timeout,
	}
}

func (p *ExecProvider) Type() string { return "plugin" }

func (p *ExecProvider) resolveExecPath(spec source.Spec) (string, error) {
	if path, ok := spec.Config["plugin_path"].(string); ok && path != "" {
		logging.Infof("plugin resolve path override source=%s path=%s", spec.Key, path)
		return path, nil
	}
	name, ok := spec.Config["plugin"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("source %q: 'plugin' or 'plugin_path' required in config", spec.Key)
	}
	path := filepath.Join(p.pluginsDir, name)
	logging.Infof("plugin resolve path source=%s plugin=%s path=%s", spec.Key, name, path)
	return path, nil
}

func (p *ExecProvider) checkExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin %q not found", path)
		}
		return fmt.Errorf("plugin %q: %w", path, err)
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("plugin %q is not executable", path)
	}
	return nil
}

func (p *ExecProvider) Validate(spec source.Spec) error {
	logging.Infof("plugin validate begin source=%s", spec.Key)
	execPath, err := p.resolveExecPath(spec)
	if err != nil {
		logging.Errorf("plugin validate resolve path failed source=%s err=%v", spec.Key, err)
		return err
	}
	if err := p.checkExecutable(execPath); err != nil {
		logging.Errorf("plugin validate executable check failed source=%s path=%s err=%v", spec.Key, execPath, err)
		return err
	}

	timeout := defaultValidateTimeout
	if p.timeout > 0 && p.timeout < timeout {
		timeout = p.timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := p.run(ctx, execPath, "validate", spec.Config)
	if err != nil {
		logging.Errorf("plugin validate run failed source=%s path=%s err=%v", spec.Key, execPath, err)
		return err
	}

	var resp PluginValidateResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return fmt.Errorf("invalid validate response: %w", err)
	}
	if !resp.Valid {
		if resp.Error != "" {
			logging.Errorf("plugin validate returned invalid source=%s path=%s err=%s", spec.Key, execPath, resp.Error)
			return fmt.Errorf("validation failed: %s", resp.Error)
		}
		logging.Errorf("plugin validate returned invalid source=%s path=%s", spec.Key, execPath)
		return fmt.Errorf("validation failed")
	}
	logging.Infof("plugin validate complete source=%s path=%s", spec.Key, execPath)
	return nil
}

func (p *ExecProvider) Fetch(ctx context.Context, spec source.Spec) (source.FetchResult, error) {
	logging.Infof("plugin fetch begin source=%s", spec.Key)
	execPath, err := p.resolveExecPath(spec)
	if err != nil {
		logging.Errorf("plugin fetch resolve path failed source=%s err=%v", spec.Key, err)
		return source.FetchResult{}, err
	}
	if err := p.checkExecutable(execPath); err != nil {
		logging.Errorf("plugin fetch executable check failed source=%s path=%s err=%v", spec.Key, execPath, err)
		return source.FetchResult{}, err
	}

	fetchCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		fetchCtx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	result, err := p.run(fetchCtx, execPath, "fetch", spec.Config)
	if err != nil {
		logging.Errorf("plugin fetch run failed source=%s path=%s err=%v", spec.Key, execPath, err)
		return source.FetchResult{}, err
	}

	var resp PluginFetchResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return source.FetchResult{},
			fmt.Errorf("invalid fetch response: %w", err)
	}
	if resp.Error != "" {
		logging.Errorf("plugin fetch returned error source=%s path=%s err=%s", spec.Key, execPath, resp.Error)
		return source.FetchResult{}, fmt.Errorf("fetch failed: %s", resp.Error)
	}

	articles := make([]domain.Article, 0, len(resp.Articles))
	for _, pa := range resp.Articles {
		var publishedAt time.Time
		if pa.PublishedAt != "" {
			parsed, err := time.Parse(time.RFC3339, pa.PublishedAt)
			if err == nil {
				publishedAt = parsed.UTC()
			}
		}
		articles = append(articles, domain.Article{
			SourceKey:   spec.Key,
			ExternalID:  pa.ExternalID,
			Title:       pa.Title,
			Link:        pa.Link,
			Author:      pa.Author,
			Summary:     pa.Summary,
			Content:     pa.Content,
			PublishedAt: publishedAt,
		})
	}

	logging.Infof("plugin fetch complete source=%s path=%s articles=%d", spec.Key, execPath, len(articles))
	return source.FetchResult{Articles: articles}, nil
}

func (p *ExecProvider) run(ctx context.Context, execPath, action string, config map[string]interface{}) ([]byte, error) {
	filteredConfig := make(map[string]interface{}, len(config))
	for k, v := range config {
		if !strings.HasPrefix(k, "plugin_") {
			filteredConfig[k] = v
		}
	}

	input, err := json.Marshal(filteredConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	cmd := exec.CommandContext(ctx, execPath, action)
	logging.Infof("plugin exec source action=%s path=%s", action, execPath)
	cmd.Stdin = bytes.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timed out after %v", p.timeout)
		}
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("exec error: %s (stderr: %s)", err, stderrStr)
		}
		return nil, fmt.Errorf("exec error: %w", err)
	}

	return stdout.Bytes(), nil
}
