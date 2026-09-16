package app

import (
	"fmt"
	"path/filepath"

	"github.com/zerodotfive/crabcore/internal/paths"
)

type InstallerSystemd struct {
	binInstallPath     string
	systemdServiceFile string
	systemdTimerFile   string
}

func NewInstallerSystemd() *InstallerSystemd {
	systemdDir := paths.SystemdUserDir()

	return &InstallerSystemd{
		binInstallPath:     paths.BinInstallPath(),
		systemdServiceFile: filepath.Join(systemdDir, "crabcore.service"),
		systemdTimerFile:   filepath.Join(systemdDir, "crabcore.timer"),
	}
}

func (inst *InstallerSystemd) GetBinInstallPath() string {
	return inst.binInstallPath
}

func (inst *InstallerSystemd) Ensure() error {
	if err := ensureBin(inst.binInstallPath); err != nil {
		return fmt.Errorf("installerSystemd: ensureBin: %s", err)
	}

	if err := ensureSystemdFile(inst.systemdServiceFile, getSystemdServiceContents(inst.binInstallPath)); err != nil {
		return fmt.Errorf("installerSystemd: ensureSystemdFile: %s", err)
	}

	if err := ensureSystemdFile(inst.systemdTimerFile, getSystemdTimerContents()); err != nil {
		return fmt.Errorf("installerSystemd: ensureSystemdFile: %s", err)
	}

	if err := ensureShellRC(); err != nil {
		return fmt.Errorf("installerSystemd: ensureShellRC: %s", err)
	}

	return nil
}
