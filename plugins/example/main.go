package main

import (
	"fmt"

	"github.com/zerodotfive/crabcore/pkg/pluginapi"

	"github.com/spf13/cobra"
)

var pluginName = "example"

var CrabcorePlugin plugin
var ExampleModule exampleModule

var _ pluginapi.SymbolProvider = (*plugin)(nil)
var _ pluginapi.Module = (*exampleModule)(nil)

var symbols = []string{"ExampleModule"}

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

type exampleModule struct{}

func (m *exampleModule) GetName() string {
	return "example"
}

func (m *exampleModule) Init(config []byte) error {
	return nil
}

func (m *exampleModule) Commands() (*cobra.Command, error) {
	return &cobra.Command{
		Use:   m.GetName(),
		Short: "Example crabcore command",
		//Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := m.run()
			return fmt.Errorf("[%s %s] %s", pluginName, m.GetName(), err)
		},
	}, nil
}

func (m *exampleModule) run() (string, error) {
	return "ok", nil
}
