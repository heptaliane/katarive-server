package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gorm.io/gorm"

	"github.com/hashicorp/go-hclog"
	"github.com/heptaliane/katarive-server/internal/service"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	"github.com/heptaliane/katarive-server/internal/service/source"
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
) (narrator.NarrateRegistry, error) {
	ctx := context.Background()

	var narrators []narrator.NarratorManager
	for _, client := range plugins.GetNarrators() {
		narrator, err := narrator.NewSemaphoreNarratorManager(ctx, client)
		if err != nil {
			return nil, err
		}
		narrators = append(narrators, narrator)
	}

	return narrator.NewFileNarratorRegistry(ctx, destDir, narrators), nil
}
func NewSource(
	interval int,
	database *gorm.DB,
	plugins *service.PluginRegistry,
) (source.SourceRegistry, error) {
	ctx := context.Background()

	var sources []source.SourceManager
	for _, rawSource := range plugins.GetSources() {
		source, err := source.NewSemaphoreSourceManager(
			ctx,
			rawSource,
			source.WithInterval(interval),
		)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	return source.NewDatabaseSourceRegistry(database, sources), nil
}
