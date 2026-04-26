package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/heptaliane/katarive-server/internal/service"
)

func LoadPlugins(pluginDir string, logLevel hclog.Level) (*service.PluginRegistry, error) {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	pr := service.NewPluginRegistry(logLevel)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(pluginDir, file.Name())
		err := pr.Load(path)
		if err != nil {
			return nil, err
		}
	}
	return pr, nil
}
func NewNarrator(
	destDir string,
	plugins *service.PluginRegistry,
) (service.NarratorRegistry, error) {
	ctx := context.Background()

	var narrators []service.NarratorManager
	for _, rawNarrator := range plugins.GetNarrators() {
		narrator, err := service.NewSemaphoreNarratorManager(ctx, rawNarrator)
		if err != nil {
			return nil, err
		}
		narrators = append(narrators, narrator)
	}

	return service.NewFileNarratorRegistry(destDir, narrators), nil
}
func NewSource(
	destDir string,
	interval int,
	plugins *service.PluginRegistry,
) (service.SourceRegistry, error) {
	ctx := context.Background()

	var sources []service.SourceManager
	for _, rawSource := range plugins.GetSources() {
		source, err := service.NewSemaphoreSourceManager(
			ctx,
			rawSource,
			service.WithInterval(interval),
		)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	return service.NewFileSourceRegistry(destDir, sources), nil
}
