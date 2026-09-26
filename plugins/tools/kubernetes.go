package main

import (
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

type kubernetesModule struct {
	crabcoreBin string
}

func (m *kubernetesModule) GetName() string {
	return "kubernetes"
}

func (m *kubernetesModule) IsRootAllowed() bool {
	return false
}

func (m *kubernetesModule) Init(config []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	m.crabcoreBin = exe

	return nil
}

func (m *kubernetesModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: "kubernetes commands",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return nil
		},
	}

	moduleRoot.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "alias for kubernetes install crabcore command",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return syscall.Exec(m.crabcoreBin, []string{m.crabcoreBin, "kubernetes", "install"}, os.Environ())
		},
	})

	return moduleRoot, nil
}
