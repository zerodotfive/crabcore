package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/zerodotfive/crabcore/internal/paths"
	"github.com/zerodotfive/crabcore/pkg/fetch"
)

type Installer interface {
	Ensure() error
	GetBinInstallPath() string
}

func NewInstaller() (Installer, error) {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/1/comm")
		if err == nil && strings.TrimSpace(string(data)) == "systemd" {
			return NewInstallerSystemd(), nil
		}
		return NewInstallerLinux(), nil
	}

	if runtime.GOOS == "darwin" {
		return NewInstallerDarwin(), nil
	}

	return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func ensureBin(binInstallPath string) error {
	currentExecutable, err := os.Executable()
	if err != nil {
		return err
	}

	if currentExecutable == binInstallPath {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(binInstallPath), 0700); err != nil {
		return err
	}

	changed, err := fetch.EnsureFile(currentExecutable, binInstallPath, false, 0755)
	if err != nil {
		return err
	}

	if err := os.Remove(currentExecutable); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if changed {
		return syscall.Exec(binInstallPath, os.Args, os.Environ())
	}

	return nil
}

func ensureShellRC() error {
	shell := path.Base(os.Getenv("SHELL"))
	switch shell {
	case "bash":
		return ensureBashrc()
	case "zsh":
		return ensureZshrc()
	default:
		return nil
	}
}

func ensureBashrc() error {
	crabcoreLines := []byte(`
PATH="$HOME/.local/bin:$PATH"
eval "$(crabcore completion bash)"
`)

	bashrcPath := paths.Bashrc()

	content, _ := os.ReadFile(bashrcPath)
	if content == nil {
		content = []byte("")
	}

	if bytes.Contains(content, crabcoreLines) {
		return nil
	}

	newContent := append(content, crabcoreLines...)
	return os.WriteFile(bashrcPath, newContent, 0644)
}

func ensureZshrc() error {
	crabcoreLines := []byte(`
PATH="$HOME/.local/bin:$PATH"
eval "source <(crabcore completion zsh)"
`)

	zshrcPath := paths.Zshrc()

	content, _ := os.ReadFile(zshrcPath)
	if content == nil {
		content = []byte("")
	}

	if bytes.Contains(content, crabcoreLines) {
		return nil
	}

	newContent := append(content, crabcoreLines...)
	return os.WriteFile(zshrcPath, newContent, 0644)
}
