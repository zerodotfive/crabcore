package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
)

func selfInstall() error {
	if os.Getuid() > 0 {
		inst, err := app.NewInstaller()
		if err != nil {
			return err
		}

		selfInstaller = &inst
		return (*selfInstaller).Ensure()
	}

	logger.L.Warn("skipping self install, because running as root")

	return nil
}

func init() {
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbosity", "v", "verbosity level")
	verbosityFlag := pflag.NewFlagSet("verbosity", pflag.ContinueOnError)
	verbosityFlag.ParseErrorsAllowlist.UnknownFlags = true
	verbosityFlag.AddFlag(rootCmd.PersistentFlags().Lookup("verbosity"))
	_ = verbosityFlag.Parse(os.Args)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("crabcore version %s\n", version.CrabcoreVersion)
		},
	})
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	logger.SetVerbosity(verbosity)

	if err := selfInstall(); err != nil {
		logger.L.Error(err.Error())
		os.Exit(1)
	}

	cfg = &config.LocalConfig{}

	if err := cfg.Read(); err != nil {
		logger.L.Warn(err.Error())
		return
	}

	if err := pluginmanager.LoadAll(cfg, rootCmd, os.Getuid() > 0); err != nil {
		logger.L.Error(err.Error())
		os.Exit(1)
	}
}

func Execute() error {
	return rootCmd.Execute()
}
