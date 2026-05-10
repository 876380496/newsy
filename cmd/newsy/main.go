package main

import (
	"flag"
	"fmt"
	"log"

	"newsy/internal/app"
	"newsy/internal/logging"
	"newsy/internal/runtimepaths"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	printPaths := flag.Bool("print-paths", false, "print runtime paths and exit")
	flag.Parse()

	paths, err := runtimepaths.Resolve(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := paths.Ensure(); err != nil {
		log.Fatal(err)
	}

	if *printPaths {
		fmt.Printf("config=%s\n", paths.ConfigFile)
		fmt.Printf("plugins=%s\n", paths.PluginDir)
		fmt.Printf("db=%s\n", paths.DBFile)
		fmt.Printf("log=%s\n", paths.LogFile)
		fmt.Printf("lock=%s\n", paths.LockFile)
		fmt.Printf("defaults=%s\n", paths.DefaultsDir)
		return
	}

	if err := logging.Init(paths.LogFile); err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = logging.Close()
	}()

	logging.Infof("application starting")
	application, err := app.New(paths)
	if err != nil {
		logging.Errorf("app.New failed: %v", err)
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		logging.Errorf("application run failed: %v", err)
		log.Fatal(err)
	}
	logging.Infof("application exited cleanly")
}
