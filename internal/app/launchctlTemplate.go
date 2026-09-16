package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func getLaunchAgentContents(binInstallPath string) []byte {
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.github.zerodotfive.crabcore</string>

    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>update</string>
    </array>

    <key>StartInterval</key>
    <integer>1800</integer>
</dict>
</plist>
`, binInstallPath))
}

func ensureLaunchAgentFile(path string, contents []byte) error {
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
		return fmt.Errorf("ensureLaunchAgentFile, WriteFile: %s", err)
	}

	if err := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%d", os.Getuid()), path).Run(); err != nil {
		return fmt.Errorf("ensureLaunchAgentFile, bootstrap: %s", err)
	}

	return nil
}
