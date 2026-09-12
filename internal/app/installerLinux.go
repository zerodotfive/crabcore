package app

import (
	"fmt"

	"github.com/zerodotfive/crabcore/internal/paths"
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

	if err := ensureBashrc(); err != nil {
		return fmt.Errorf("installerLinux: ensureBashrc: %s", err)
	}

	return nil
}
