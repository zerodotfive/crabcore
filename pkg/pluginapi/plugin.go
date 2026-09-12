package pluginapi

import "github.com/spf13/cobra"

type SymbolProvider interface {
	GetRoot() (*cobra.Command, error)
	GetSymbols() []string
}
