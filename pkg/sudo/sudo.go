package sudo

import (
	"os"
	"os/exec"
	"syscall"
)

func EnsureSudo() error {
	if os.Getuid() > 0 {
		exe, err := os.Executable()
		if err != nil {
			return err
		}

		sudoPath, err := exec.LookPath("sudo")
		if err != nil {
			return err
		}

		return syscall.Exec(sudoPath, append([]string{sudoPath, exe}, os.Args[1:]...), os.Environ())
	}

	return nil
}
