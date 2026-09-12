package cmd

import (
	"fmt"
	"os"

	"github.com/zerodotfive/crabcore/internal/app"
	"github.com/zerodotfive/crabcore/internal/config"
	"github.com/zerodotfive/crabcore/internal/pluginmanager"
	"github.com/zerodotfive/crabcore/internal/version"

	"github.com/spf13/cobra"
)

var (
	selfInstaller *app.Installer
	cfg           *config.LocalConfig
	rootCmd       = &cobra.Command{
		Use:     "crabcore",
		Short:   "Workspace tools configurator",
		Long:    "crabcore is a workspace tools configurator",
		Version: version.CrabcoreVersion,
		CompletionOptions: cobra.CompletionOptions{
			DisableDescriptions: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return nil
		},
	}
)

func init() {
	inst, err := app.NewInstaller()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	selfInstaller = &inst

	if err := (*selfInstaller).Ensure(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	cfg = &config.LocalConfig{}
	if err := cfg.Read(); err != nil {
		return
	}

	if err := pluginmanager.Load(cfg, rootCmd); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func Execute() error {
	return rootCmd.Execute()
}
