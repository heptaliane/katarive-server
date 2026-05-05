package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/heptaliane/katarive-server/internal/service"
)

const DEFAULT_PLUGIN_PREFIX string = "default"

func LoadPlugins(pluginDir string, logLevel hclog.Level) (*service.PluginRegistry, error) {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(files, func(a, b os.DirEntry) int {
		nameA := a.Name()
		nameB := b.Name()
		hasPrefixA := strings.HasPrefix(nameA, DEFAULT_PLUGIN_PREFIX)
		hasPrefixB := strings.HasPrefix(nameB, DEFAULT_PLUGIN_PREFIX)

		if hasPrefixA && !hasPrefixB {
			return 1
		}
		if !hasPrefixA && hasPrefixB {
			return -1
		}
		return strings.Compare(nameA, nameB)
	})

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
