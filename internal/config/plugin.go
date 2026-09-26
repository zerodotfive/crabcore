package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/internal/pluginloader"
	"github.com/zerodotfive/crabcore/pkg/fetch"
	"github.com/zerodotfive/crabcore/pkg/logger"
	"github.com/zerodotfive/crabcore/pkg/paths"
	"github.com/zerodotfive/crabcore/pkg/pluginapi"
)

type Plugin struct {
	Name         string            `yaml:"plugin"`
	Filename     string            `yaml:"filename"`
	URL          map[string]string `yaml:"url"`
	ModuleConfig map[string]string `yaml:"moduleConfig"`
	loader       *pluginloader.Loader
	rootCommand  *cobra.Command
	modules      map[string]pluginapi.Module
}

func (plugin *Plugin) loadModule(module pluginapi.Module) error {
	moduleConfigPath := paths.ModuleConfigPath(plugin.Name, module.GetName())
	moduleConfig, err := fetch.Fetch(moduleConfigPath, false)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.Getuid() == 0 && !module.IsRootAllowed() {
		logger.L.Warn(fmt.Sprintf("module %s in plugin %s is not allowed to run as root, skipping", module.GetName(), plugin.Name))
		return nil
	}

	if err := module.Init(moduleConfig); err != nil {
		return err
	}
	moduleRoot, _ := module.Commands()

	plugin.rootCommand.AddCommand(moduleRoot)

	return nil
}

func (plugin *Plugin) LoadAllModules() error {
	if err := plugin.getPluginModules(); err != nil {
		return err
	}

	for _, module := range plugin.modules {
		if err := plugin.loadModule(module); err != nil {
			return err
		}
	}
	return nil
}

func (plugin *Plugin) getPluginLoader() (*pluginloader.Loader, error) {
	if plugin.loader != nil {
		return plugin.loader, nil
	}

	pluginLibPath := paths.PluginLibPath(plugin.Filename)
	loader, err := pluginloader.NewLoader(pluginLibPath)
	if err != nil {
		return nil, err
	}

	plugin.loader = loader

	return plugin.loader, nil
}

func (plugin *Plugin) getPluginModules() error {
	loader, err := plugin.getPluginLoader()
	if err != nil {
		return err
	}

	modules, err := loader.GetContents()
	if err != nil {
		return err
	}

	plugin.modules = modules

	return nil
}

func (plugin *Plugin) GetPluginRootCommand() (*cobra.Command, error) {
	loader, err := plugin.getPluginLoader()
	if err != nil {
		return nil, err
	}

	pluginRoot, err := loader.GetRootSymbol()
	if err != nil {
		return nil, err
	}

	pluginRootCommand, err := pluginRoot()
	if err != nil {
		return nil, err
	}

	plugin.rootCommand = pluginRootCommand

	return pluginRootCommand, nil
}
