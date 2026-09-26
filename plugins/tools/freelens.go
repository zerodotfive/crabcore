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

var freelensUbuntuAptSourceTemplate = `Types: deb
URIs: https://github.com/freelensapp/freelens/releases/latest/download
Suites: ./
Signed-By: /etc/apt/keyrings/freelens.asc
`

var freelensAnyErrorTemplate = "try to check output in more verbose mode with -vvv"

type freelensModule struct {
	LensDMGUrl map[string]string `yaml:"lensDMGUrl"`
	LensRPMUrl map[string]string `yaml:"lensRPMUrl"`
}

func (m *freelensModule) GetName() string {
	return "freelens"
}

func (m *freelensModule) IsRootAllowed() bool {
	return true
}

func (m *freelensModule) Init(config []byte) error {
	err := yaml.Unmarshal(config, m)
	if err != nil {
		return err
	}

	return nil
}

func (m *freelensModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: "freelens commands",
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
		Short: "Install freelens",
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

			r, err := m.freelensInstall()
			if err != nil {
				fmt.Printf("[freelens install] %s\n", r)
				return err
			}

			fmt.Printf("[freelens install] %s\n", r)
			return nil
		},
	})

	return moduleRoot, nil
}

func (m *freelensModule) freelensInstall() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return m.freelensInstallMacos()
	case "linux":
		return m.freelensInstallLinux()
	default:
		return "", errors.New("os not supported")
	}
}

func (m *freelensModule) freelensInstallLinux() (string, error) {
	rel, err := osrelease.GetOSRelease()
	if err != nil {
		return "", err
	}

	switch rel.ID {
	case "debian", "ubuntu":
		return m.freelensInstallDebian()
	case "fedora":
		return m.freelensInstallFedora()
	default:
		return "", errors.New("os not supported")
	}
}

func (m *freelensModule) freelensInstallDebian() (string, error) {
	if _, err := fetch.EnsureFile("https://raw.githubusercontent.com/freelensapp/freelens/refs/heads/main/freelens/build/apt/freelens.asc", "/etc/apt/keyrings/freelens.asc", 0644); err != nil {
		return "", err
	}

	if err := os.WriteFile("/etc/apt/sources.list.d/freelens.sources", []byte(fmt.Sprintf(freelensUbuntuAptSourceTemplate)), 0644); err != nil {
		return "", err
	}

	if err := run.Cmd("apt", "update"); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	if err := run.Cmd("apt", "install", "-y", "freelens"); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	return "Installation complete", nil
}

func (m *freelensModule) freelensInstallFedora() (string, error) {
	url, ok := m.LensRPMUrl[runtime.GOARCH]
	if !ok {
		return "", errors.New("no freelens url for arch found in config")
	}

	tmp, err := os.CreateTemp(paths.CrabcoreTMPDir(), "freelens-*.rpm")
	if err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	_ = tmp.Close()
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := fetch.EnsureFileWithProgress(url, tmp.Name(), 0644); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	if err := run.Cmd("dnf", "install", "-y", tmp.Name()); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	return "Installation complete", nil
}

func (m *freelensModule) freelensInstallMacos() (string, error) {
	url, ok := m.LensDMGUrl[runtime.GOARCH]
	if !ok {
		return "", errors.New("no freelens url for arch found in config")
	}

	tmp, err := os.CreateTemp(paths.CrabcoreTMPDir(), "freelens-*.dmg")
	if err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	_ = tmp.Close()
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := fetch.EnsureFileWithProgress(url, tmp.Name(), 0644); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	if err := run.Cmd("hdiutil attach " + tmp.Name()); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	if err := run.Cmd("cp -a /Volumes/freelens/Freelens.app /Applications/Freelens.app"); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	if err := run.Cmd("hdiutil detach " + tmp.Name()); err != nil {
		return fmt.Sprintf(freelensAnyErrorTemplate), err
	}

	return "Installation complete", nil
}
