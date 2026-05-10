package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AvailablePlugins lists the executable files in pluginDir.
// Non-executable files, directories, and hidden files are skipped.
// If the directory does not exist, nil is returned (not an error).
func AvailablePlugins(pluginDir string) ([]string, error) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read plugin dir %s: %w", pluginDir, err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0o111 == 0 {
			continue
		}

		typeName := strings.TrimSuffix(name, filepath.Ext(name))
		names = append(names, typeName)
	}

	return names, nil
}
