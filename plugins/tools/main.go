package main

import (
	"fmt"

	"github.com/zerodotfive/crabcore/pkg/pluginapi"

	"github.com/spf13/cobra"
)

var pluginName = "tools"

var CrabcorePlugin plugin
var PowerlineGoModule powerlineGoModule
var KubernetesModule kubernetesModule
var DockerModule dockerModule
var LensappModule lensappModule
var FreelensModule freelensModule

var _ pluginapi.SymbolProvider = (*plugin)(nil)
var _ pluginapi.Module = (*powerlineGoModule)(nil)
var _ pluginapi.Module = (*kubernetesModule)(nil)
var _ pluginapi.Module = (*dockerModule)(nil)
var _ pluginapi.Module = (*lensappModule)(nil)
var _ pluginapi.Module = (*freelensModule)(nil)

var symbols = []string{"PowerlineGoModule", "KubernetesModule", "DockerModule", "LensappModule", "FreelensModule"}

type plugin struct{}

func (p *plugin) GetRoot() (*cobra.Command, error) {
	pluginRoot := &cobra.Command{
		Use:   pluginName,
		Short: fmt.Sprintf("Tools commands"),
	}

	return pluginRoot, nil
}

func (p *plugin) GetSymbols() []string {
	return symbols
}
