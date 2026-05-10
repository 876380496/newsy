package runtimepaths

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed defaults/config.yaml
var embeddedDefaultConfig []byte

type Paths struct {
	ConfigFile     string
	ConfigDir      string
	PluginDir      string
	DataDir        string
	DBFile         string
	CacheDir       string
	LogFile        string
	LockFile       string
	DefaultsDir    string
	ConfigExplicit bool
}

func Resolve(configOverride string) (Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve home dir: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve executable path: %w", err)
	}
	if evalPath, evalErr := filepath.EvalSymlinks(execPath); evalErr == nil {
		execPath = evalPath
	}

	defaultConfigDir := envOr("NEWSY_CONFIG_DIR", filepath.Join(homeDir, ".config", "newsy"))
	defaultDataDir := envOr("NEWSY_DATA_DIR", filepath.Join(homeDir, ".local", "share", "newsy"))
	defaultCacheDir := envOr("NEWSY_CACHE_DIR", filepath.Join(homeDir, ".cache", "newsy"))

	configFile := configOverride
	configExplicit := configOverride != ""
	if configFile == "" {
		configFile = os.Getenv("NEWSY_CONFIG")
		configExplicit = configFile != ""
	}
	if configFile == "" {
		configFile = filepath.Join(defaultConfigDir, "config.yaml")
	}

	pluginDir := envOr("NEWSY_PLUGIN_DIR", filepath.Join(defaultConfigDir, "plugins"))
	dbFile := envOr("NEWSY_DB_FILE", filepath.Join(defaultDataDir, "newsy.db"))
	logFile := envOr("NEWSY_LOG_FILE", filepath.Join(defaultCacheDir, "newsy.log"))
	lockFile := envOr("NEWSY_LOCK_FILE", filepath.Join(defaultCacheDir, "newsy.lock"))
	defaultsDir := envOr("NEWSY_DEFAULTS_DIR", filepath.Clean(filepath.Join(filepath.Dir(execPath), "..", "share", "newsy")))

	configFile, err = filepath.Abs(configFile)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve config path: %w", err)
	}
	pluginDir, err = filepath.Abs(pluginDir)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve plugin dir: %w", err)
	}
	dbFile, err = filepath.Abs(dbFile)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve db path: %w", err)
	}
	logFile, err = filepath.Abs(logFile)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve log path: %w", err)
	}
	lockFile, err = filepath.Abs(lockFile)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve lock path: %w", err)
	}
	defaultsDir, err = filepath.Abs(defaultsDir)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve defaults dir: %w", err)
	}

	return Paths{
		ConfigFile:     configFile,
		ConfigDir:      filepath.Dir(configFile),
		PluginDir:      pluginDir,
		DataDir:        filepath.Dir(dbFile),
		DBFile:         dbFile,
		CacheDir:       filepath.Dir(logFile),
		LogFile:        logFile,
		LockFile:       lockFile,
		DefaultsDir:    defaultsDir,
		ConfigExplicit: configExplicit,
	}, nil
}

func (p Paths) Ensure() error {
	for _, dir := range []string{p.ConfigDir, p.PluginDir, p.DataDir, filepath.Dir(p.LogFile), filepath.Dir(p.LockFile)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	if p.ConfigExplicit {
		return nil
	}
	if _, err := os.Stat(p.ConfigFile); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat config file %s: %w", p.ConfigFile, err)
	}

	data, err := p.defaultConfig()
	if err != nil {
		return err
	}
	if err := os.WriteFile(p.ConfigFile, data, 0o644); err != nil {
		return fmt.Errorf("write default config %s: %w", p.ConfigFile, err)
	}
	return nil
}

func (p Paths) defaultConfig() ([]byte, error) {
	defaultConfigPath := filepath.Join(p.DefaultsDir, "config.yaml")
	if data, err := os.ReadFile(defaultConfigPath); err == nil {
		return data, nil
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read default config %s: %w", defaultConfigPath, err)
	}
	return embeddedDefaultConfig, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
