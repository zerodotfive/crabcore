package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/fetch"
	"github.com/zerodotfive/crabcore/pkg/paths"
	"gopkg.in/yaml.v3"
)

var powerlineGoLinuxFontconfigFileContents = []byte(`<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
  <alias>
    <family>monospace</family>
    <prefer>
      <family>PowerlineSymbols</family>
    </prefer>
  </alias>
</fontconfig>`)

var powerlineGoMacOSInstructions = []byte(`Install one of
NERD FONTS https://www.nerdfonts.com/font-downloads
or
POWERLINE FONTS https://github.com/powerline/fonts
then put it into ~/Library/Fonts directory and select it in your terminal settings`)

type powerlineGoModule struct {
	crabcoreBin        string
	powerlineGoPath    string
	PowerlineGoVersion string            `yaml:"powerlineGoVersion"`
	PowerlineGoShellRC map[string]string `yaml:"powerlineGoShellRC"`
}

func (m *powerlineGoModule) GetName() string {
	return "powerline-go"
}

func (m *powerlineGoModule) IsRootAllowed() bool {
	return false
}

func (m *powerlineGoModule) Init(config []byte) error {
	err := yaml.Unmarshal(config, m)
	if err != nil {
		return err
	}

	m.powerlineGoPath = path.Join(paths.BinDir(), "powerline-go")

	return nil
}

func (m *powerlineGoModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: "powerline-go commands",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		//Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return nil
		},
	}

	moduleRoot.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install powerline-go",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := m.powerlineGo()
			if err != nil {
				return err
			}

			fmt.Printf("[powerline-go install] %s\n", r)

			return nil
		},
	})

	return moduleRoot, nil
}

func getPowerlineGoSHA256SumByOSArch(filename string, version string) (string, error) {
	remoteSHA256SumsUrl := fmt.Sprintf("https://github.com/justjanne/powerline-go/releases/download/%s/checksums.txt", version)
	remoteSHA256Sums, err := fetch.Fetch(remoteSHA256SumsUrl, false)
	if err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(string(remoteSHA256Sums), "\n") {
		fields := strings.Fields(line)
		if fields[1] == filename {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("could not find checksum for %s", filename)
}

func (m *powerlineGoModule) powerlineGoShellRC() error {
	shell := path.Base(os.Getenv("SHELL"))
	rcFile := ""
	if rcPart, ok := m.PowerlineGoShellRC[shell]; ok {
		switch shell {
		case "bash":
			rcFile = path.Join(paths.Home(), ".bashrc")
		case "zsh":
			rcFile = path.Join(paths.Home(), ".zshrc")
		default:
		}

		content, _ := os.ReadFile(rcFile)
		if content == nil {
			content = []byte("")
		}

		if bytes.Contains(content, []byte(rcPart)) {
			return nil
		}

		newContent := append(content, rcPart...)
		err := os.WriteFile(rcFile, newContent, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *powerlineGoModule) powerlineGoFonts() error {
	switch runtime.GOOS {
	case "darwin":
		fmt.Printf("%s\n", powerlineGoMacOSInstructions)
		return nil
	case "linux":
		fontconfigDir := path.Join(paths.Home(), ".config/fontconfig/conf.d")
		fontconfigFile := path.Join(fontconfigDir, "10-powerline-symbols.conf")
		if err := os.MkdirAll(fontconfigDir, 0755); err != nil {
			return err
		}
		if _, err := os.Stat(fontconfigFile); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(fontconfigFile, powerlineGoLinuxFontconfigFileContents, 0644); err != nil {
				return err
			}
		}

		openTypeDir := path.Join(paths.Home(), ".local/share/fonts/opentype")
		fontFileUrl := "https://github.com/powerline/powerline/raw/refs/heads/develop/font/PowerlineSymbols.otf"
		fontFilePath := path.Join(openTypeDir, "PowerlineSymbols.otf")
		if err := os.MkdirAll(openTypeDir, 0755); err != nil {
			return err
		}
		if _, err := os.Stat(fontFilePath); errors.Is(err, os.ErrNotExist) {
			if _, err = fetch.EnsureFileWithProgress(fontFileUrl, fontFilePath, 0644); err != nil {
				return err
			}
		}

		return nil
	default:
		return errors.New("didn't know how to install powerline fonts")
	}
}

func (m *powerlineGoModule) powerlineGo() (string, error) {
	filename := "powerline-go-" + runtime.GOOS + "-" + runtime.GOARCH
	remoteSHA256Sum, err := getPowerlineGoSHA256SumByOSArch(filename, m.PowerlineGoVersion)
	if err != nil {
		return "", err
	}

	localSHA256Sum, err := fetch.GetSHA256(m.powerlineGoPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if localSHA256Sum != remoteSHA256Sum {
		url := fmt.Sprintf("https://github.com/justjanne/powerline-go/releases/download/%s/powerline-go-linux-amd64", m.PowerlineGoVersion)
		_, err = fetch.EnsureFileWithProgress(url, m.powerlineGoPath, 0755)
		if err != nil {
			return "", err
		}
	}

	if err := m.powerlineGoShellRC(); err != nil {
		return "", err
	}

	if err := m.powerlineGoFonts(); err != nil {
		return "", err
	}

	return "ok", nil
}
