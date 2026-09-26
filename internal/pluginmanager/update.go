package pluginmanager

import (
	"fmt"
	"runtime"

	"github.com/zerodotfive/crabcore/internal/config"
	"github.com/zerodotfive/crabcore/pkg/paths"
)

func Update(cfg *config.LocalConfig) (bool, error) {
	changed := false
	if cfg.Cache == nil {
		return false, nil
	}

	for _, plugin := range cfg.Cache.Plugins {
		pluginLibPath := paths.PluginLibPath(plugin.Filename)

		var err error
		osarch := runtime.GOOS + "-" + runtime.GOARCH
		if _, ok := plugin.URL[osarch]; !ok {
			return false, fmt.Errorf("no plugin url for platform %s found", osarch)
		}
		tmpChanged, err := ensureLib(pluginLibPath, plugin.URL[osarch], true)
		changed = changed || tmpChanged
		if err != nil {
			return false, err
		}
	}

	return changed, nil
}
