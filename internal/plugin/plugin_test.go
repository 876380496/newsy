package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"newsy/internal/source"
)

func TestAvailablePlugins_SkipsNonexecutable(t *testing.T) {
	dir := t.TempDir()
	nonExec := filepath.Join(dir, "readonly.sh")
	if err := os.WriteFile(nonExec, []byte("#!/bin/sh\necho hi"), 0644); err != nil {
		t.Fatal(err)
	}

	names, err := AvailablePlugins(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 plugins for non-executable file, got %d", len(names))
	}
}

func TestAvailablePlugins_DirNotExist(t *testing.T) {
	names, err := AvailablePlugins("/nonexistent/plugins")
	if err != nil {
		t.Fatal(err)
	}
	if names != nil {
		t.Fatalf("expected nil for nonexistent dir, got %d plugins", len(names))
	}
}

func TestAvailablePlugins_SkipsHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	hidden := filepath.Join(dir, ".hidden.sh")
	if err := os.WriteFile(hidden, []byte("#!/bin/sh\necho hi"), 0755); err != nil {
		t.Fatal(err)
	}

	names, err := AvailablePlugins(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 plugins for hidden file, got %d", len(names))
	}
}

func TestAvailablePlugins_FindsExecutable(t *testing.T) {
	dir := t.TempDir()
	plugin := filepath.Join(dir, "myscraper")
	if err := os.WriteFile(plugin, []byte("#!/bin/sh\necho hi"), 0755); err != nil {
		t.Fatal(err)
	}

	names, err := AvailablePlugins(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(names))
	}
	if names[0] != "myscraper" {
		t.Fatalf("expected 'myscraper', got %q", names[0])
	}
}

func TestType(t *testing.T) {
	ep := NewExecProvider("./plugins")
	if ep.Type() != "plugin" {
		t.Fatalf("expected 'plugin', got %q", ep.Type())
	}
}

func TestExecProvider_Validate_Valid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "testplugin"), []byte(`#!/bin/sh
if [ "$1" = "validate" ]; then
  echo '{"valid":true}'
fi`), 0755); err != nil {
		t.Fatal(err)
	}

	ep := NewExecProvider(dir)
	err := ep.Validate(source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "testplugin", "url": "https://example.com"},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestExecProvider_Validate_Invalid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "testplugin"), []byte(`#!/bin/sh
if [ "$1" = "validate" ]; then
  echo '{"valid":false,"error":"url is required"}'
fi`), 0755); err != nil {
		t.Fatal(err)
	}

	ep := NewExecProvider(dir)
	err := ep.Validate(source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "testplugin"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecProvider_Fetch_Success(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "testplugin"), []byte(`#!/bin/sh
if [ "$1" = "fetch" ]; then
  cat <<'JSON'
{"articles":[
  {"external_id":"a1","title":"Article 1","link":"https://example.com/1"},
  {"external_id":"a2","title":"Article 2","link":"https://example.com/2"}
]}
JSON
fi`), 0755); err != nil {
		t.Fatal(err)
	}

	ep := NewExecProvider(dir)
	result, err := ep.Fetch(context.Background(), source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "testplugin"},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result.Articles) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(result.Articles))
	}
	if result.Articles[0].SourceKey != "test" {
		t.Fatalf("expected SourceKey 'test', got %q", result.Articles[0].SourceKey)
	}
	if result.Articles[0].ExternalID != "a1" {
		t.Fatalf("expected ExternalID 'a1', got %q", result.Articles[0].ExternalID)
	}
}

func TestExecProvider_Fetch_FullArticle(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "testplugin"), []byte(`#!/bin/sh
if [ "$1" = "fetch" ]; then
  cat <<'JSON'
{"articles":[
  {
    "external_id":"a1","title":"Full Article","link":"https://example.com/1",
    "author":"Author","summary":"Summary text","content":"Full content",
    "published_at":"2026-05-10T12:00:00Z"
  }
]}
JSON
fi`), 0755); err != nil {
		t.Fatal(err)
	}

	ep := NewExecProvider(dir)
	result, err := ep.Fetch(context.Background(), source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "testplugin"},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	a := result.Articles[0]
	if a.Title != "Full Article" || a.Author != "Author" || a.Summary != "Summary text" || a.Content != "Full content" {
		t.Fatalf("unexpected article fields: %+v", a)
	}
	if !a.PublishedAt.Equal(time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected PublishedAt: %v", a.PublishedAt)
	}
}

func TestExecProvider_Fetch_Error(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "testplugin"), []byte(`#!/bin/sh
if [ "$1" = "fetch" ]; then
  echo '{"error":"something went wrong"}'
fi`), 0755); err != nil {
		t.Fatal(err)
	}

	ep := NewExecProvider(dir)
	_, err := ep.Fetch(context.Background(), source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "testplugin"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecProvider_Fetch_NonZeroExit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "testplugin"), []byte(`#!/bin/sh
echo "error message" >&2
exit 1`), 0755); err != nil {
		t.Fatal(err)
	}

	ep := NewExecProvider(dir)
	_, err := ep.Fetch(context.Background(), source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "testplugin"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveExecPath_PluginName(t *testing.T) {
	ep := NewExecProvider("/some/dir")
	path, err := ep.resolveExecPath(source.Spec{
		Key:    "test",
		Config: map[string]interface{}{"plugin": "myscript"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/some/dir/myscript" {
		t.Fatalf("expected /some/dir/myscript, got %q", path)
	}
}

func TestResolveExecPath_PluginPathOverride(t *testing.T) {
	ep := NewExecProvider("/some/dir")
	path, err := ep.resolveExecPath(source.Spec{
		Key: "test",
		Config: map[string]interface{}{
			"plugin":      "myscript",
			"plugin_path": "/custom/path",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/custom/path" {
		t.Fatalf("expected /custom/path, got %q", path)
	}
}

func TestResolveExecPath_Missing(t *testing.T) {
	ep := NewExecProvider("/some/dir")
	_, err := ep.resolveExecPath(source.Spec{
		Key:    "test",
		Config: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("expected error for missing plugin config, got nil")
	}
}
