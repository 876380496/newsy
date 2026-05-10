package runtimepaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUsesConfigOverride(t *testing.T) {
	override := filepath.Join(t.TempDir(), "custom.yaml")
	paths, err := Resolve(override)
	if err != nil {
		t.Fatal(err)
	}
	if paths.ConfigFile != override {
		t.Fatalf("expected config override %q, got %q", override, paths.ConfigFile)
	}
	if !paths.ConfigExplicit {
		t.Fatal("expected config override to be explicit")
	}
}

func TestResolveDefaultsToXDGStyleDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("NEWSY_CONFIG_DIR", "")
	t.Setenv("NEWSY_DATA_DIR", "")
	t.Setenv("NEWSY_CACHE_DIR", "")
	t.Setenv("NEWSY_PLUGIN_DIR", "")
	t.Setenv("NEWSY_DB_FILE", "")
	t.Setenv("NEWSY_LOG_FILE", "")
	t.Setenv("NEWSY_LOCK_FILE", "")

	paths, err := Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if paths.ConfigFile != filepath.Join(home, ".config", "newsy", "config.yaml") {
		t.Fatalf("unexpected config path: %s", paths.ConfigFile)
	}
	if paths.PluginDir != filepath.Join(home, ".config", "newsy", "plugins") {
		t.Fatalf("unexpected plugin dir: %s", paths.PluginDir)
	}
	if paths.DBFile != filepath.Join(home, ".local", "share", "newsy", "newsy.db") {
		t.Fatalf("unexpected db path: %s", paths.DBFile)
	}
	if paths.LogFile != filepath.Join(home, ".cache", "newsy", "newsy.log") {
		t.Fatalf("unexpected log path: %s", paths.LogFile)
	}
}

func TestEnsureCreatesDefaultConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("NEWSY_CONFIG_DIR", filepath.Join(root, "config"))
	t.Setenv("NEWSY_DATA_DIR", filepath.Join(root, "data"))
	t.Setenv("NEWSY_CACHE_DIR", filepath.Join(root, "cache"))
	t.Setenv("NEWSY_PLUGIN_DIR", filepath.Join(root, "plugins"))
	t.Setenv("NEWSY_DEFAULTS_DIR", filepath.Join(root, "defaults"))

	paths, err := Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(paths.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected default config to be written")
	}
}

func TestEnsureRespectsDefaultsDirConfig(t *testing.T) {
	root := t.TempDir()
	defaultsDir := filepath.Join(root, "defaults")
	if err := os.MkdirAll(defaultsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := []byte("sources:\n  - key: test\n    provider_type: rss\n")
	if err := os.WriteFile(filepath.Join(defaultsDir, "config.yaml"), want, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("NEWSY_CONFIG_DIR", filepath.Join(root, "config"))
	t.Setenv("NEWSY_DATA_DIR", filepath.Join(root, "data"))
	t.Setenv("NEWSY_CACHE_DIR", filepath.Join(root, "cache"))
	t.Setenv("NEWSY_PLUGIN_DIR", filepath.Join(root, "plugins"))
	t.Setenv("NEWSY_DEFAULTS_DIR", defaultsDir)

	paths, err := Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(paths.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("expected config from defaults dir, got %q", string(got))
	}
}
