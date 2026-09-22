package fetch

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func zipToTMP(destDir string, zippedFile *zip.File) (string, error) {
	f, err := zippedFile.Open()
	if err != nil {
		return "", err
	}

	defer func() { _ = f.Close() }()

	fromZipTmp, err := os.CreateTemp(destDir, ".crabcore-*")
	if err != nil {
		return "", fmt.Errorf("%s: %s", zippedFile.Name, err)
	}

	defer func() { _ = fromZipTmp.Close() }()

	if _, err := io.Copy(fromZipTmp, f); err != nil {
		return "", err
	}

	return fromZipTmp.Name(), nil
}

func EnsureFileFromZipWithSHA256WithProgress(url string, fileNameInZip string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
	destDir := filepath.Dir(dest)

	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	var archive *zip.ReadCloser
	defer func() { _ = archive.Close() }()

	tmp, err := os.CreateTemp(destDir, ".crabcore-*")
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	defer func() { _ = os.Remove(tmp.Name()) }()

	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	_, err = EnsureFileWithProgress(url, tmp.Name(), mode)
	if err != nil {
		return false, fmt.Errorf("tmp: %s: %s", tmp.Name(), err)
	}

	archive, err = zip.OpenReader(tmp.Name())
	if err != nil {
		return false, fmt.Errorf("zip: %s: %s", tmp.Name(), err)
	}

	zipTmpFilename := ""
	defer func() { _ = os.Remove(zipTmpFilename) }()

	for _, zippedFile := range archive.File {

		if zippedFile.Name != fileNameInZip {
			continue
		}

		tmpFromZipFilename, err := zipToTMP(destDir, zippedFile)
		if err != nil {
			return false, fmt.Errorf("%s: %s", zippedFile.Name, err)
		}

		if err := os.Chmod(tmpFromZipFilename, mode); err != nil {
			return false, fmt.Errorf("%s: %s", dest, err)
		}

		if err := os.Rename(tmpFromZipFilename, dest); err != nil {
			return false, fmt.Errorf("%s: %s", dest, err)
		}
	}

	return false, err
}
