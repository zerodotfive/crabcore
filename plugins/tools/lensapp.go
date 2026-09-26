package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/fetch"
	"github.com/zerodotfive/crabcore/pkg/osrelease"
	"github.com/zerodotfive/crabcore/pkg/paths"
	"github.com/zerodotfive/crabcore/pkg/run"
	"github.com/zerodotfive/crabcore/pkg/sudo"
	"gopkg.in/yaml.v3"
)

var lensappUbuntuAptSourceTemplate = `Types: deb
URIs: https://downloads.k8slens.dev/apt/debian
Suites: stable
Components: main
Architectures: %s
Signed-By: /etc/apt/keyrings/lens.asc
`

var lensappAnyErrorTemplate = "try to check output in more verbose mode with -vvv"

type lensappModule struct {
	LensDMGUrl map[string]string `yaml:"lensDMGUrl"`
}

func (m *lensappModule) GetName() string {
	return "lensapp"
}

func (m *lensappModule) IsRootAllowed() bool {
	return true
}

func (m *lensappModule) Init(config []byte) error {
	err := yaml.Unmarshal(config, m)
	if err != nil {
		return err
	}

	return nil
}

func (m *lensappModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: "lensapp commands",
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
		Short: "Install lensapp",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(paths.CrabcoreTMPDir(), 0755); err != nil {
				return err
			}

			if err := sudo.EnsureSudo(); err != nil {
				return err
			}

			r, err := m.lensappInstall()
			if err != nil {
				fmt.Printf("[lensapp install] %s\n", r)
				return err
			}

			fmt.Printf("[lensapp install] %s\n", r)
			return nil
		},
	})

	return moduleRoot, nil
}

func (m *lensappModule) lensappInstall() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return m.lensappInstallMacos()
	case "linux":
		return m.lensappInstallLinux()
	default:
		return "", errors.New("os not supported")
	}
}

func (m *lensappModule) lensappInstallLinux() (string, error) {
	rel, err := osrelease.GetOSRelease()
	if err != nil {
		return "", err
	}

	switch rel.ID {
	case "debian", "ubuntu":
		return m.lensappInstallDebian()
	case "fedora":
		return m.lensappInstallFedora()
	default:
		return "", errors.New("os not supported")
	}
}

func (m *lensappModule) lensappInstallDebian() (string, error) {
	if _, err := fetch.EnsureFile("https://downloads.k8slens.dev/keys/gpg", "/etc/apt/keyrings/lens.asc", 0644); err != nil {
		return "", err
	}

	if err := os.WriteFile("/etc/apt/sources.list.d/lens.sources", []byte(fmt.Sprintf(lensappUbuntuAptSourceTemplate, runtime.GOARCH)), 0644); err != nil {
		return "", err
	}

	if err := run.Cmd("apt", "update"); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	if err := run.Cmd("apt", "install", "-y", "lens"); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	return "Installation complete", nil
}

func (m *lensappModule) lensappInstallFedora() (string, error) {
	if err := run.Cmd("dnf", "config-manager", "addrepo", "--overwrite", "--from-repofile", "https://downloads.k8slens.dev/rpm/lens.repo"); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	if err := run.Cmd("dnf", "install", "-y", "lens"); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	return "Installation complete", nil
}

func (m *lensappModule) lensappInstallMacos() (string, error) {
	url, ok := m.LensDMGUrl[runtime.GOARCH]
	if !ok {
		return "", errors.New("no lensapp url for arch found in config")
	}

	tmp, err := os.CreateTemp(paths.CrabcoreTMPDir(), "lensapp-*.dmg")
	if err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	_ = tmp.Close()
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := fetch.EnsureFileWithProgress(url, tmp.Name(), 0644); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	if err := run.Cmd("hdiutil attach " + tmp.Name()); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	if err := run.Cmd("cp -a /Volumes/lensapp/Lens.app /Applications/Lens.app"); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	if err := run.Cmd("hdiutil detach " + tmp.Name()); err != nil {
		return fmt.Sprintf(lensappAnyErrorTemplate), err
	}

	return "Installation complete", nil
}
