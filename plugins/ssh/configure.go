package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/paths"
	"gopkg.in/yaml.v3"
)

type sshConfigdEntry struct {
	Filename string `yaml:"filename"`
	Content  string `yaml:"content"`
}
type sshConfigureModule struct {
	sshConfigPath     string
	sshConfigdPath    string
	sshConfigdEntries []sshConfigdEntry
}

func (m *sshConfigureModule) GetName() string {
	return "configure"
}

func (m *sshConfigureModule) IsRootAllowed() bool {
	return false
}

func (m *sshConfigureModule) Init(config []byte) error {
	if len(config) == 0 {
		return errors.New("empty config")
	}

	m.sshConfigPath = path.Join(paths.Home(), ".ssh/config")
	m.sshConfigdPath = path.Join(paths.Home(), ".ssh/crabcore.d")

	return yaml.Unmarshal(config, &m.sshConfigdEntries)
}

func (m *sshConfigureModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: fmt.Sprintf("(Re)configure %s, %s from plugin cache", m.sshConfigPath, m.sshConfigdPath),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		Annotations: map[string]string{
			"run-on-update": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := m.run()
			if err != nil {
				return fmt.Errorf("[%s %s] %s", pluginName, m.GetName(), err)
			}

			fmt.Printf("[%s %s] %s\n", pluginName, m.GetName(), r)

			return nil
		},
	}

	return moduleRoot, nil
}

func (m *sshConfigureModule) run() (string, error) {
	_, err := exec.LookPath("ssh")
	if err != nil {
		return "", fmt.Errorf("ssh not found in $PATH, please install it first")
	}

	if err := os.MkdirAll(m.sshConfigdPath, 0700); err != nil {
		return "", err
	}

	crabcoreLines := []byte("# crabcore\nInclude crabcore.d/*\n")

	content, err := os.ReadFile(m.sshConfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		content = []byte("")
	}

	if !bytes.Contains(content, crabcoreLines) {
		newContent := append(content, crabcoreLines...)
		if err := os.WriteFile(m.sshConfigPath, newContent, 0644); err != nil {
			return "", err
		}
	}

	for _, entry := range m.sshConfigdEntries {
		f := path.Join(m.sshConfigdPath, entry.Filename)
		err = os.WriteFile(f, []byte(entry.Content), 0600)
		if err != nil {
			return "", err
		}
	}

	return "ok", nil
}
