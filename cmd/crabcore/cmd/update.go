package cmd

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/zerodotfive/crabcore/internal/app"
	"github.com/zerodotfive/crabcore/internal/config"
	"github.com/zerodotfive/crabcore/internal/pluginmanager"

	"github.com/spf13/cobra"
)

var (
	cfgURL string
)

func init() {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update config and plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			changed := false
			tmpChanged := false

			if cfgURL == "" {
				if cfg.URL == "" {
					return fmt.Errorf(
						"no config url found in config or commandline.\n" +
							"Run `crabcore update --config <config url or filename>`\n",
					)
				}

				cfgURL = cfg.URL
			}

			var err error
			tmpChanged, err = config.Update(cfg, cfgURL, true)
			changed = changed || tmpChanged
			if err != nil {
				return err
			}

			osarch := runtime.GOOS + "-" + runtime.GOARCH
			if _, ok := cfg.Cache.BinURL[osarch]; !ok {
				return fmt.Errorf("no bin url platform: %s", osarch)
			}
			tmpChanged, err = app.Update(cfg.Cache.BinURL[osarch], (*selfInstaller).GetBinInstallPath())
			changed = changed || tmpChanged
			if err != nil {
				return err
			}

			tmpChanged, err = pluginmanager.Update(cfg)
			changed = changed || tmpChanged
			if err != nil {
				return err
			}

			if changed {
				return syscall.Exec((*selfInstaller).GetBinInstallPath(), os.Args, os.Environ())
			}

			return runUpdateAnnotatedCommands(rootCmd)
		},
	}

	updateCmd.Flags().StringVarP(
		&cfgURL,
		"config",
		"c",
		"",
		"Config url or filename",
	)

	rootCmd.AddCommand(updateCmd)
}

func runUpdateAnnotatedCommands(cmd *cobra.Command) error {
	if value, ok := cmd.Annotations["run-on-update"]; ok && value == "true" {
		if err := cmd.RunE(cmd, nil); err != nil {
			return fmt.Errorf("command %s failed: %s", cmd.Name(), err)
		}
	}

	for _, c := range cmd.Commands() {
		if err := runUpdateAnnotatedCommands(c); err != nil {
			return err
		}
	}

	return nil
}
