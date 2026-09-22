package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/internal/app"
	"github.com/zerodotfive/crabcore/internal/config"
	"github.com/zerodotfive/crabcore/internal/pluginmanager"
	"github.com/zerodotfive/crabcore/internal/version"
	"github.com/zerodotfive/crabcore/pkg/logger"
)

var (
	selfInstaller *app.Installer
	cfg           *config.LocalConfig
	verbosity     int
	rootCmd       = &cobra.Command{
		Use:   "crabcore",
		Short: "Workspace tools configurator",
		Long:  "crabcore is a workspace tools configurator",
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
		logger.L.Error(err.Error())
		os.Exit(1)
	}

	selfInstaller = &inst

	if err := (*selfInstaller).Ensure(); err != nil {
		logger.L.Error(err.Error())
		os.Exit(1)
	}

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbosity", "v", "verbosity level")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("crabcore version %s\n", version.CrabcoreVersion)
		},
	})

	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	cobra.OnInitialize(func() {
		logger.SetVerbose(verbosity)
	})

	cfg = &config.LocalConfig{}
	if err := cfg.Read(); err != nil {
		return
	}

	if err := pluginmanager.Load(cfg, rootCmd); err != nil {
		logger.L.Error(err.Error())
		os.Exit(1)
	}
}

func Execute() error {
	return rootCmd.Execute()
}
