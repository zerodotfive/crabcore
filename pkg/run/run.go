package run

import (
	"os/exec"

	"github.com/zerodotfive/crabcore/pkg/logger"
)

func Cmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	logger.L.Debug(string(out))
	if err != nil {
		logger.L.Debug(string(out))
		return err
	}
	return nil
}
