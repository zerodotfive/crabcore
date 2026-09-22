package fetch

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

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

func EnsureFileFromTarSHA256WithProgress(url string, fileNameInTar string, dest string, remoteSHAFileSuffix string, mode fs.FileMode) (bool, error) {
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

	if _, err := EnsureFileWithProgress(url, archivePath, mode); err != nil {
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
