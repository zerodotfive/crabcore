package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/fetch"
	"gopkg.in/yaml.v3"
)

type installModule struct {
	kubectlUrl       string
	kubectlPath      string
	KubectlVersion   string `yaml:"kubectlVersion"`
	kubeloginUrl     string
	kubeloginPath    string
	KubeloginVersion string            `yaml:"kubeloginVersion"`
	KubeloginSHA256  map[string]string `yaml:"kubeloginSHA256"`
	helmUrl          string
	helmPath         string
	HelmVersion      string            `yaml:"helmVersion"`
	HelmSHA256       map[string]string `yaml:"helmSHA256"`
}

func (m *installModule) GetName() string {
	return "install"
}

func (m *installModule) Init(config []byte) error {
	err := yaml.Unmarshal(config, m)
	if err != nil {
		return err
	}

	m.kubectlUrl = fmt.Sprintf("https://dl.k8s.io/release/%s/bin/%s/%s/kubectl", m.KubectlVersion, runtime.GOOS, runtime.GOARCH)
	m.kubectlPath = path.Join(os.Getenv("HOME"), ".local/bin", "kubectl")

	m.kubeloginUrl = fmt.Sprintf("https://github.com/int128/kubelogin/releases/download/%s/kubelogin_%s_%s.zip", m.KubeloginVersion, runtime.GOOS, runtime.GOARCH)
	m.kubeloginPath = path.Join(os.Getenv("HOME"), ".local/bin", "kubectl-oidc_login")

	m.helmUrl = fmt.Sprintf("https://get.helm.sh/helm-%s-%s-%s.tar.gz", m.HelmVersion, runtime.GOOS, runtime.GOARCH)
	m.helmPath = path.Join(os.Getenv("HOME"), ".local/bin", "helm")

	return nil
}

func (m *installModule) Commands() (*cobra.Command, error) {
	return &cobra.Command{
		Use:   m.GetName(),
		Short: "Install kubernetes tools",
		//Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := m.run()
			if err != nil {
				return fmt.Errorf("[%s %s] %s", pluginName, m.GetName(), err)
			}

			fmt.Printf("[%s %s] %s\n", pluginName, m.GetName(), r)

			return nil
		},
	}, nil
}

func (m *installModule) run() (string, error) {
	_, err := fetch.EnsureFile(m.kubectlUrl, m.kubectlPath, ".sha256", 0755)
	if err != nil {
		return "", err
	}

	kubeloginLocalSHA256, err := fetch.GetSHA256(m.kubeloginPath)
	if err != nil {
		return "", err
	}

	if kubeloginLocalSHA256 != m.KubeloginSHA256[runtime.GOOS+"-"+runtime.GOARCH] {
		_, err = fetch.EnsureFileFromZipSHA256(m.kubeloginUrl, "kubelogin", m.kubeloginPath, ".sha256", 0755)
		if err != nil {
			return "", err
		}
	}

	helmLocalSHA256, err := fetch.GetSHA256(m.kubeloginPath)
	if err != nil {
		return "", err
	}

	if helmLocalSHA256 != m.HelmSHA256[runtime.GOOS+"-"+runtime.GOARCH] {
		_, err = fetch.EnsureFileFromTarSHA256(m.helmUrl, runtime.GOOS+"-"+runtime.GOARCH+"/helm", m.helmPath, ".sha256sum", 0755)
		if err != nil {
			return "", err
		}
	}

	return "ok", nil
}
