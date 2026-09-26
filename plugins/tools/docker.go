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
)

var dockerUbuntuAptSourceTemplate = `Types: deb
URIs: https://download.docker.com/linux/%s
Suites: %s
Components: stable
Architectures: %s
Signed-By: /etc/apt/keyrings/docker.asc
`

var dockerInstallErrorTemplate = "If docker previously installed try to check output in more verbose mode with -vvv. Or run `%s`"
var dockerAnyErrorTemplate = "try to check output in more verbose mode with -vvv"
var dockerLogoutTemplate = "Installation complete. Logout and login back to use `docker` command without sudo"

type dockerModule struct{}

func (m *dockerModule) GetName() string {
	return "docker"
}

func (m *dockerModule) IsRootAllowed() bool {
	return true
}

func (m *dockerModule) Init(config []byte) error {
	return nil
}

func (m *dockerModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: "docker commands",
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
		Short: "Install docker",
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

			r, err := m.dockerInstall()
			if err != nil {
				fmt.Printf("[docker install] %s\n", r)
				return err
			}

			fmt.Printf("[docker install] %s\n", r)
			return nil
		},
	})

	return moduleRoot, nil
}

func (m *dockerModule) dockerInstall() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return m.dockerInstallMacos()
	case "linux":
		return m.dockerInstallLinux()
	default:
		return "", errors.New("os not supported")
	}
}

func (m *dockerModule) dockerInstallLinux() (string, error) {
	rel, err := osrelease.GetOSRelease()
	if err != nil {
		return "", err
	}

	switch rel.ID {
	case "debian", "ubuntu":
		return m.dockerInstallDebian(rel)
	case "fedora":
		return m.dockerInstallFedora(rel)
	default:
		return "", errors.New("os not supported")
	}
}

func (m *dockerModule) dockerInstallDebian(rel *osrelease.OSRelease) (string, error) {
	if _, err := fetch.EnsureFile(fmt.Sprintf("https://download.docker.com/linux/%s/gpg", rel.ID), "/etc/apt/keyrings/docker.asc", 0644); err != nil {
		return "", err
	}

	if err := os.WriteFile("/etc/apt/sources.list.d/docker.sources", []byte(fmt.Sprintf(dockerUbuntuAptSourceTemplate, rel.ID, rel.VersionCodename, runtime.GOARCH)), 0644); err != nil {
		return "", err
	}

	if err := run.Cmd("apt", "update"); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	if err := run.Cmd("apt", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"); err != nil {
		return fmt.Sprintf(dockerInstallErrorTemplate, "sudo apt remove $(dpkg --get-selections docker.io docker-compose docker-compose-v2 docker-doc docker-buildx podman-docker containerd runc | cut -f1)"), err
	}

	if err := run.Cmd("usermod", "-aG", "docker", os.Getenv("SUDO_USER")); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	return fmt.Sprintf(dockerLogoutTemplate), nil
}

func (m *dockerModule) dockerInstallFedora(rel *osrelease.OSRelease) (string, error) {
	if err := run.Cmd("dnf", "config-manager", "addrepo", "--overwrite", "--from-repofile", "https://download.docker.com/linux/fedora/docker-ce.repo"); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	if err := run.Cmd("dnf", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"); err != nil {
		return fmt.Sprintf(dockerInstallErrorTemplate, "sudo dnf remove docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-selinux docker-engine-selinux docker-engine"), err
	}

	if err := run.Cmd("usermod", "-aG", "docker", os.Getenv("SUDO_USER")); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	return fmt.Sprintf(dockerLogoutTemplate), nil
}

func (m *dockerModule) dockerInstallMacos() (string, error) {
	tmp, err := os.CreateTemp(paths.CrabcoreTMPDir(), ".crabcore-*.dmg")
	if err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	_ = tmp.Close()
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := fetch.EnsureFileWithProgress(fmt.Sprintf("https://desktop.docker.com/mac/main/%s/Docker.dmg", runtime.GOARCH), tmp.Name(), 0644); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	if err := run.Cmd("hdiutil attach " + tmp.Name()); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	if err := run.Cmd("/Volumes/Docker/Docker.app/Contents/MacOS/install --accept-license --user=" + os.Getenv("SUDO_USER")); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	if err := run.Cmd("hdiutil detach " + tmp.Name()); err != nil {
		return fmt.Sprintf(dockerAnyErrorTemplate), err
	}

	return "Installation complete", nil
}
