package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("sources:\n  - key: hn\n    provider_type: rss\n    name: Hacker News\n    enabled: true\n    config:\n      url: https://hnrss.org/frontpage\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Key != "hn" {
		t.Fatalf("unexpected source key: %s", cfg.Sources[0].Key)
	}
}
