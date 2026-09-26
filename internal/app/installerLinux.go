package app

import (
	"fmt"

	"github.com/zerodotfive/crabcore/pkg/paths"
)

type InstallerLinux struct {
	binInstallPath string
}

func NewInstallerLinux() *InstallerLinux {
	return &InstallerLinux{
		binInstallPath: paths.BinInstallPath(),
	}
}

func (inst *InstallerLinux) GetBinInstallPath() string {
	return inst.binInstallPath
}

func (inst *InstallerLinux) Ensure() error {
	if err := ensureBin(inst.binInstallPath); err != nil {
		return fmt.Errorf("installerLinux: ensureBin: %s", err)
	}

	if err := ensureShellRC(); err != nil {
		return fmt.Errorf("installerLinux: ensureShellRC: %s", err)
	}

	return nil
}
