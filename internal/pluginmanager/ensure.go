package pluginmanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zerodotfive/crabcore/pkg/fetch"
)

func ensureLib(pluginPath string, url string, forceUpdate bool) (bool, error) {
	libDir := filepath.Dir(pluginPath)

	if _, err := os.Stat(pluginPath); err == nil && !forceUpdate {
		return false, nil
	}

	if err := os.MkdirAll(libDir, 0755); err != nil {
		return false, fmt.Errorf("%s: %s", libDir, err)
	}

	return fetch.EnsureFile(url, pluginPath, true, 0755)
}
