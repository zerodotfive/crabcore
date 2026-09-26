package fetch

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/zerodotfive/crabcore/pkg/logger"
)

func EnsureFile(url string, dest string, mode fs.FileMode) (bool, error) {
	return ensureFile(url, dest, "", mode, false)
}

func EnsureFileWithSHA256(url string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
	return ensureFile(url, dest, remoteSHAFileSuffix, mode, false)
}

func EnsureFileWithProgress(url string, dest string, mode fs.FileMode) (bool, error) {
	return ensureFile(url, dest, "", mode, true)
}

func EnsureFileWithSHA256WithProgress(url string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
	return ensureFile(url, dest, remoteSHAFileSuffix, mode, true)
}

func ensureFile(url string, dest string, remoteSHAFileSuffix string, mode fs.FileMode, progress bool) (bool, error) {
	destDir := filepath.Dir(dest)
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	var fileData []byte
	var remoteSHA256 string

	var localSHA256 string
	destStat, err := os.Stat(dest)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	if destStat != nil && !os.IsNotExist(err) && destStat.Size() != 0 {
		localSHA256, err = GetSHA256(dest)
		if err != nil {
			return false, fmt.Errorf("%s: %s", dest, err)
		}
	}

	if remoteSHAFileSuffix != "" {
		remoteSHA256byte, err := Fetch(url+remoteSHAFileSuffix, false)
		if err != nil {
			logger.L.Info(fmt.Sprintf("will try to get full file sum %s, because of remote sum fetch error: %s", url, err.Error()))
		}

		remoteSHA256 = strings.TrimSpace(string(remoteSHA256byte))
		fields := strings.Fields(remoteSHA256)
		if len(fields) > 1 {
			remoteSHA256 = fields[0]
		}

		if localSHA256 != "" && localSHA256 == remoteSHA256 {
			return false, nil
		}

		if err == nil {
			logger.L.Info(fmt.Sprintf("will try to get full file sum %s, because of local '%s' and remote '%s' is not equal", url, localSHA256, remoteSHA256))
		}
	}

	fileData, err = Fetch(url, progress)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	remoteHash := sha256.Sum256(fileData)
	remoteSHA256 = hex.EncodeToString(remoteHash[:])

	if localSHA256 == remoteSHA256 {
		return false, nil
	}

	logger.L.Info(fmt.Sprintf("will update %s, because of SHA old '%s' new '%s'", dest, localSHA256, remoteSHA256))

	tmp, err := os.CreateTemp(destDir, ".crabcore-*")
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(fileData); err != nil {
		_ = tmp.Close()
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	if err := os.Chmod(tmp.Name(), mode); err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	if err := os.Rename(tmp.Name(), dest); err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	return true, nil
}
