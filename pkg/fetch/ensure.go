package fetch

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func EnsureFile(url string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
	destDir := filepath.Dir(dest)

	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	var fileData []byte
	var remoteSHA256 string

	localSHA256, err := GetSHA256(dest)
	if err != nil {
		return false, fmt.Errorf("%s: %s", dest, err)
	}

	if remoteSHAFileSuffix != "" {
		remoteSHA256byte, err := Fetch(url + remoteSHAFileSuffix)
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

func EnsureFileFromZipSHA256(url string, fileNameInZip string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
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

	_, err = EnsureFile(url, tmp.Name(), remoteSHAFileSuffix, mode)
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

func newTarReader(r io.Reader) (*tar.Reader, func() error, error) {
	br := bufio.NewReader(r)

	magic, err := br.Peek(2)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, nil, err
	}

	if len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		gz, err := gzip.NewReader(br)
		if err != nil {
			return nil, nil, err
		}

		return tar.NewReader(gz), gz.Close, nil
	}

	return tar.NewReader(br), func() error { return nil }, nil
}

func tarToTMP(destDir string, r io.Reader) (_ string, err error) {
	f, err := os.CreateTemp(destDir, ".crabcore-*")
	if err != nil {
		return "", err
	}

	tmpName := f.Name()

	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = io.Copy(f, r); err != nil {
		_ = f.Close()
		return "", err
	}

	if err = f.Close(); err != nil {
		return "", err
	}

	return tmpName, nil
}

func cleanTarName(name string) string {
	return path.Clean("/" + name)
}

func EnsureFileFromTarSHA256(url string, fileNameInTar string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
	destDir := filepath.Dir(dest)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return false, fmt.Errorf("%s: %w", dest, err)
	}

	tmp, err := os.CreateTemp(destDir, ".crabcore-*")
	if err != nil {
		return false, fmt.Errorf("%s: %w", dest, err)
	}

	archivePath := tmp.Name()

	defer func() { _ = os.Remove(archivePath) }()

	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("%s: %w", dest, err)
	}

	if _, err := EnsureFile(url, archivePath, remoteSHAFileSuffix, mode); err != nil {
		return false, fmt.Errorf("tmp: %s: %w", archivePath, err)
	}

	archive, err := os.Open(archivePath)
	if err != nil {
		return false, fmt.Errorf("tar: %s: %w", archivePath, err)
	}

	defer func() { _ = archive.Close() }()

	tr, closeDecompressor, err := newTarReader(archive)
	if err != nil {
		return false, fmt.Errorf("tar: %s: %w", archivePath, err)
	}

	defer func() { _ = closeDecompressor() }()

	want := cleanTarName(fileNameInTar)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return false, fmt.Errorf("tar: %s: %w", archivePath, err)
		}

		if hdr.Typeflag != tar.TypeReg || cleanTarName(hdr.Name) != want {
			continue
		}

		tmpFromTar, err := tarToTMP(destDir, tr)
		if err != nil {
			return false, fmt.Errorf("%s: %w", hdr.Name, err)
		}

		if err := os.Chmod(tmpFromTar, mode); err != nil {
			_ = os.Remove(tmpFromTar)
			return false, fmt.Errorf("%s: %w", dest, err)
		}

		if err := os.Rename(tmpFromTar, dest); err != nil {
			_ = os.Remove(tmpFromTar)
			return false, fmt.Errorf("%s: %w", dest, err)
		}

		return true, nil
	}

	return false, fmt.Errorf("%s: not found in %s", fileNameInTar, url)
}
