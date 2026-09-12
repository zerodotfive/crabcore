package app

import (
	"fmt"

	"github.com/zerodotfive/crabcore/internal/paths"
)

type InstallerDarwin struct {
	binInstallPath string
}

func NewInstallerDarwin() *InstallerDarwin {
	return &InstallerDarwin{
		binInstallPath: paths.BinInstallPath(),
	}
}

func (inst *InstallerDarwin) GetBinInstallPath() string {
	return inst.binInstallPath
}

func (inst *InstallerDarwin) Ensure() error {
	if err := ensureBin(inst.binInstallPath); err != nil {
		return fmt.Errorf("installerDarwin: ensureBin: %s", err)
	}

	if err := ensureBashrc(); err != nil {
		return fmt.Errorf("installerDarwin: ensureBashrc: %s", err)
	}

	return nil
}
