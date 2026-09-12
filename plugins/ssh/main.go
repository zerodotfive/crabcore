package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/pluginapi"
)

var pluginName = "ssh"

var CrabcorePlugin plugin
var SSHConfigureModule sshConfigureModule
var SSHParallelModule sshParallelModule

var _ pluginapi.SymbolProvider = (*plugin)(nil)
var _ pluginapi.Module = (*sshConfigureModule)(nil)
var _ pluginapi.Module = (*sshParallelModule)(nil)

var symbols = []string{
	"SSHConfigureModule",
	"SSHParallelModule",
}

type plugin struct{}

func (p *plugin) GetRoot() (*cobra.Command, error) {
	pluginRoot := &cobra.Command{
		Use:   pluginName,
		Short: fmt.Sprintf("SSH commands"),
	}

	return pluginRoot, nil
}

func (p *plugin) GetSymbols() []string {
	return symbols
}
