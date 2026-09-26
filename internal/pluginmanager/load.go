package pluginmanager

import (
	"debug/buildinfo"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/internal/config"
	"github.com/zerodotfive/crabcore/pkg/logger"
	"github.com/zerodotfive/crabcore/pkg/paths"
)

func loadPlugin(plugin config.Plugin, rootCmd *cobra.Command, doEnsureLib bool) error {
	pluginLibPath := paths.PluginLibPath(plugin.Filename)
	osarch := runtime.GOOS + "-" + runtime.GOARCH
	if _, ok := plugin.URL[osarch]; !ok {
		return fmt.Errorf("no plugin url for platform %s found", osarch)
	}

	if doEnsureLib {
		if _, err := ensureLib(pluginLibPath, plugin.URL[osarch], false); err != nil {
			return err
		}
	}

	pluginRoot, err := plugin.GetPluginRootCommand()
	if err != nil {
		return err
	}

	if err := plugin.LoadAllModules(); err != nil {
		return err
	}

	if len(pluginRoot.Commands()) == 0 {
		logger.L.Warn(fmt.Sprintf("no modules loaded for plugin %s, skipping plugin", plugin.Name))
		return nil
	}

	rootCmd.AddCommand(pluginRoot)

	return nil
}

func checkPluginGoVersionConflict(pluginFilename string) (error, bool) {
	pluginLibPath := paths.PluginLibPath(pluginFilename)
	if _, err := os.Stat(pluginLibPath); err != nil {
		return nil, false
	}

	info, err := buildinfo.ReadFile(pluginLibPath)
	if err != nil {
		return err, true
	}

	return nil, info.GoVersion != runtime.Version()
}

func LoadAll(cfg *config.LocalConfig, rootCmd *cobra.Command, doEnsureLib bool) error {
	if cfg.Cache == nil {
		return nil
	}

	for _, plugin := range cfg.Cache.Plugins {
		if err, conflict := checkPluginGoVersionConflict(plugin.Filename); err != nil || conflict {
			logger.L.Warn("plugin " + plugin.Name + " go runtime version conflict, skipping plugins load")
			break
		}

		if err := loadPlugin(plugin, rootCmd, doEnsureLib); err != nil {
			return err
		}
	}

	return nil
}
