package pluginloader

import (
	"fmt"
	"os"
	goplugin "plugin"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/logger"
	"github.com/zerodotfive/crabcore/pkg/pluginapi"
)

type Loader struct {
	plugin         *goplugin.Plugin
	symbolProvider pluginapi.SymbolProvider
}

func NewLoader(pluginPath string) (*Loader, error) {
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

			logger.L.Error(fmt.Sprintf("%s is built with a different version of package, restarting", pluginPath))
			return nil, syscall.Exec(exePath, os.Args, os.Environ())
		}

		return nil, fmt.Errorf("%s: %s", pluginPath, err)
	}

	return &Loader{
		plugin: p,
	}, nil
}

func (loader *Loader) lookup() error {
	if loader.symbolProvider != nil {
		return nil
	}

	symPlugin, err := loader.plugin.Lookup("CrabcorePlugin")
	if err != nil {
		return err
	}

	s, ok := symPlugin.(pluginapi.SymbolProvider)
	if !ok {
		return fmt.Errorf("unexpected type symbol in plugin")
	}

	loader.symbolProvider = s

	return nil
}

func (loader *Loader) GetRootSymbol() (func() (*cobra.Command, error), error) {
	if err := loader.lookup(); err != nil {
		return nil, err
	}

	return loader.symbolProvider.GetRoot, nil
}

func (loader *Loader) getModulesSymbols() ([]string, error) {
	if err := loader.lookup(); err != nil {
		return nil, err
	}

	symbols := loader.symbolProvider.GetSymbols()
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols found")
	}

	return symbols, nil
}

func (loader *Loader) GetContents() (map[string]pluginapi.Module, error) {
	symbols, err := loader.getModulesSymbols()
	if err != nil {
		return nil, err
	}

	modules := make(map[string]pluginapi.Module)

	for _, symbol := range symbols {
		symMod, err := loader.plugin.Lookup(symbol)
		if err != nil {
			return nil, err
		}

		m, ok := symMod.(pluginapi.Module)
		if !ok {
			return nil, fmt.Errorf("unexpected type symbol %s in module", symbol)
		}
		modules[symbol] = m
	}

	return modules, nil
}
