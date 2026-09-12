package fetch

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func EnsureFile(url string, dest string, tryFetchSHA256 bool, mode fs.FileMode) (bool, error) {
	destDir := filepath.Dir(dest)

	var fileData []byte
	var remoteSHA256 string

	localSHA256, err := getSHA256(dest)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	if tryFetchSHA256 {
		remoteSHA256byte, err := Fetch(url + ".sha256")
		if remoteSHA256byte != nil {
			remoteSHA256 = strings.TrimSpace(string(remoteSHA256byte))
		}

		if localSHA256 != "" && localSHA256 == remoteSHA256 {
			return false, nil
		}

		if err != nil {
			fmt.Printf("Trying get full file sum %s, because of remote sum fetch error: %s\n", url, err.Error())
		} else {
			fmt.Printf("Trying get full file sum %s, because of local '%s' and remote '%s' is not equal\n", url, localSHA256, remoteSHA256)
		}
	}

	fileData, err = Fetch(url)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	remoteHash := sha256.Sum256(fileData)
	remoteSHA256 = hex.EncodeToString(remoteHash[:])

	if localSHA256 == remoteSHA256 {
		return false, nil
	}

	fmt.Printf("Will update %s, because of SHA old '%s' new '%s'\n", dest, localSHA256, remoteSHA256)

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
