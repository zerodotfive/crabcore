package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/zerodotfive/crabcore/internal/paths"
	"github.com/zerodotfive/crabcore/pkg/fetch"
)

func Update(cfg *LocalConfig, url string, forceRefetch bool) (bool, error) {
	changed := false
	if url != "" && cfg.URL != url {
		cfg.URL = url
		cfg.Cache = &RuntimeConfig{}
		changed = true
	}

	if cfg.Cache == nil || forceRefetch {
		cache := RuntimeConfig{}
		if err := cache.Read(url); err != nil {
			return false, err
		}

		if !reflect.DeepEqual(*(cfg.Cache), cache) {
			cfg.Cache = &cache
			changed = true
		}
	}

	if err := cfg.Write(); err != nil {
		return false, err
	}

	for _, plugin := range cfg.Cache.Plugins {
		pluginConfigDir := paths.PluginConfigDir(plugin.Name)
		if err := os.MkdirAll(pluginConfigDir, 0755); err != nil {
			return false, fmt.Errorf("%s: %s", pluginConfigDir, err)
		}
		for module, moduleConfigURL := range plugin.ModuleConfig {
			tmpChanged, err := fetch.EnsureFile(moduleConfigURL, paths.ModuleConfigPath(plugin.Name, module), 0600)
			changed = changed || tmpChanged
			if err != nil {
				return false, err
			}
		}
	}

	return changed, nil
}
