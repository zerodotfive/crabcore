package app

import (
	"fmt"
	"path/filepath"

	"github.com/zerodotfive/crabcore/internal/paths"
)

type InstallerDarwin struct {
	binInstallPath        string
	darwinLaunchAgentFile string
}

func NewInstallerDarwin() *InstallerDarwin {
	return &InstallerDarwin{
		binInstallPath:        paths.BinInstallPath(),
		darwinLaunchAgentFile: filepath.Join(paths.LaunchAgentDir(), "com.github.zerodotfive.crabcore.plist"),
	}
}

func (inst *InstallerDarwin) GetBinInstallPath() string {
	return inst.binInstallPath
}

func (inst *InstallerDarwin) Ensure() error {
	if err := ensureBin(inst.binInstallPath); err != nil {
		return fmt.Errorf("installerDarwin: ensureBin: %s", err)
	}

	if err := ensureLaunchAgentFile(inst.darwinLaunchAgentFile, getLaunchAgentContents(inst.binInstallPath)); err != nil {
		return fmt.Errorf("installerDarwin: ensureLaunchAgentFile: %s", err)
	}

	if err := ensureShellRC(); err != nil {
		return fmt.Errorf("installerDarwin: ensureShellRC: %s", err)
	}

	return nil
}
