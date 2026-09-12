package pluginapi

import "github.com/spf13/cobra"

type Module interface {
	GetName() string
	Init(config []byte) error
	Commands() (*cobra.Command, error)
}
