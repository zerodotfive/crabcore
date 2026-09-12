package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func getSystemdServiceContents(binInstallPath string) []byte {
	return []byte(fmt.Sprintf(`[Unit]
Description=crabcore
Wants=crabcore.timer

[Service]
Type=oneshot
ExecStart=%s update

[Install]
WantedBy=multi-user.target
`, binInstallPath))
}

func getSystemdTimerContents() []byte {
	return []byte(`[Unit]
Description=crabcore

[Timer]
OnCalendar=*:0/30
Persistent=true

[Install]
WantedBy=timers.target
`)
}

func ensureSystemdFile(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	old, err := os.ReadFile(path)
	if err == nil {
		if bytes.Equal(old, contents) {
			return nil
		}
	}

	if err := os.WriteFile(path, contents, 0600); err != nil {
		return fmt.Errorf("ensureSystemdFile, WriteFile: %s", err)
	}

	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("ensureSystemdFile, daemon-reload: %s", err)
	}

	if err := exec.Command("systemctl", "--user", "enable", filepath.Base(path)).Run(); err != nil {
		return fmt.Errorf("ensureSystemdFile, enable: %s", err)
	}

	_ = exec.Command("systemctl", "--user", "restart", filepath.Base(path)).Run()

	return nil
}
