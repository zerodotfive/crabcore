package pluginapi

import "github.com/spf13/cobra"

type Module interface {
	GetName() string
	IsRootAllowed() bool
	Init(config []byte) error
	Commands() (*cobra.Command, error)
}
