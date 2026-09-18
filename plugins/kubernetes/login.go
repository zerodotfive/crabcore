package main

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type loginModule struct {
	Clusters map[string]string `yaml:"clusters"`
}

func (m *loginModule) GetName() string {
	return "login"
}

func (m *loginModule) Init(config []byte) error {
	err := yaml.Unmarshal(config, m)
	if err != nil {
		return err
	}

	return nil
}

func (m *loginModule) Commands() (*cobra.Command, error) {
	return &cobra.Command{
		Use:   m.GetName(),
		Short: fmt.Sprintf("Login against kubernetes cluster"),
		//Hidden: true,
		Long: fmt.Sprintf("crabcore kubernetes login <cluster name>]"),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := exec.Command("bash", "-c", m.Clusters[args[0]], args[0])
			command.Env = append(os.Environ(),
				"KNAME="+args[0],
			)

			return command.Run()
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return slices.Collect(maps.Keys(m.Clusters)), cobra.ShellCompDirectiveNoFileComp
		},
	}, nil
}
