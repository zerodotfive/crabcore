package app

import (
	"os"

	"github.com/zerodotfive/crabcore/pkg/fetch"
)

func Update(url string, dest string) (bool, error) {
	currentExecutable, err := os.Executable()
	if err != nil {
		return false, err
	}

	if currentExecutable != dest {
		return true, nil
	}

	changed, err := fetch.EnsureFileWithSHA256WithProgress(url, dest, ".sha256", 0755)
	if err != nil {
		return false, err
	}

	return changed, nil
}
