package pluginmanager

import (
	"fmt"
	"os"
	goplugin "plugin"
	"runtime"
	"strings"
	"syscall"

	"github.com/zerodotfive/crabcore/internal/config"
	"github.com/zerodotfive/crabcore/internal/paths"
	"github.com/zerodotfive/crabcore/pkg/fetch"
	"github.com/zerodotfive/crabcore/pkg/pluginapi"

	"github.com/spf13/cobra"
)

type Loader struct {
	plugin *goplugin.Plugin
}

func Load(cfg *config.LocalConfig, rootCmd *cobra.Command) error {
	if cfg.Cache == nil {
		return nil
	}

	for _, plugin := range cfg.Cache.Plugins {
		pluginLibPath := paths.PluginLibPath(plugin.Filename)

		osarch := runtime.GOOS + "-" + runtime.GOARCH
		if _, ok := plugin.URL[osarch]; !ok {
			return fmt.Errorf("no plugin url for platform %s found", osarch)
		}
		if _, err := ensureLib(pluginLibPath, plugin.URL[osarch], false); err != nil {
			return err
		}

		loader, err := newLoader(pluginLibPath)
		if err != nil {
			return err
		}

		pluginRoot, modules, err := loader.getContents()
		if err != nil {
			return err
		}

		pluginRootCommand, err := pluginRoot()
		if err != nil {
			return err
		}
		rootCmd.AddCommand(pluginRootCommand)

		for _, module := range *modules {
			moduleConfigPath := paths.ModuleConfigPath(plugin.Name, module.GetName())
			moduleConfig, err := fetch.Fetch(moduleConfigPath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := module.Init(moduleConfig); err != nil {
				fmt.Printf("error loading plugin config %s: %s\n", moduleConfig, err)
				continue
			}
			moduleRoot, _ := module.Commands()

			pluginRootCommand.AddCommand(moduleRoot)
		}
	}

	return nil
}

func newLoader(pluginPath string) (*Loader, error) {
	p, err := goplugin.Open(pluginPath)
	if err != nil {
		if strings.Contains(err.Error(), "plugin was built with a different version of package") {
			if err := os.Remove(pluginPath); err != nil {
				return nil, err
			}

			exePath, err := os.Executable()
			if err != nil {
				return nil, err
			}

			fmt.Printf("%s is built with a different version of package, restarting\n", pluginPath)
			return nil, syscall.Exec(exePath, os.Args, os.Environ())
		}

		return nil, fmt.Errorf("%s: %s", pluginPath, err)
	}

	return &Loader{
		plugin: p,
	}, nil
}

func (l *Loader) getSymbols() (func() (*cobra.Command, error), []string, error) {
	symPlugin, err := l.plugin.Lookup("CrabcorePlugin")
	if err != nil {
		return nil, nil, err
	}

	s, ok := symPlugin.(pluginapi.SymbolProvider)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected type symbol in plugin")
	}

	root := s.GetRoot

	symbols := s.GetSymbols()
	if len(symbols) == 0 {
		return nil, nil, fmt.Errorf("no symbols found")
	}

	return root, symbols, nil
}

func (l *Loader) getContents() (func() (*cobra.Command, error), *map[string]pluginapi.Module, error) {
	root, symbols, err := l.getSymbols()
	if err != nil {
		return nil, nil, err
	}

	modules := make(map[string]pluginapi.Module)

	for _, symbol := range symbols {
		symMod, err := l.plugin.Lookup(symbol)
		if err != nil {
			return nil, nil, err
		}

		m, ok := symMod.(pluginapi.Module)
		if !ok {
			return nil, nil, fmt.Errorf("unexpected type symbol %s in module", symbol)
		}
		modules[symbol] = m
	}

	return root, &modules, nil
}
