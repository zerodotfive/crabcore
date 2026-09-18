package main

import (
	"fmt"

	"github.com/zerodotfive/crabcore/pkg/pluginapi"

	"github.com/spf13/cobra"
)

var pluginName = "kubernetes"

var CrabcorePlugin plugin
var InstallModule installModule
var LoginModule loginModule

var _ pluginapi.SymbolProvider = (*plugin)(nil)
var _ pluginapi.Module = (*installModule)(nil)
var _ pluginapi.Module = (*loginModule)(nil)

var symbols = []string{"InstallModule", "LoginModule"}

type plugin struct{}

func (p *plugin) GetRoot() (*cobra.Command, error) {
	pluginRoot := &cobra.Command{
		Use:   pluginName,
		Short: fmt.Sprintf("Kubernetes commands"),
	}

	return pluginRoot, nil
}

func (p *plugin) GetSymbols() []string {
	return symbols
}
