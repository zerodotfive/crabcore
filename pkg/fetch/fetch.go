package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func Fetch(url string, progress bool) ([]byte, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error loading file: %s", resp.Status)
		}

		if progress {
			progressReader := &progressReader{url, resp.Body, resp.ContentLength, 0}

			return io.ReadAll(progressReader)
		}

		return io.ReadAll(resp.Body)
	}

	filename := url
	if strings.HasPrefix(url, "file://") {
		filename = strings.TrimPrefix(url, "file:/")
	}

	if !strings.HasPrefix(filename, "/") {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		filename = cwd + "/" + filename
	}

	return os.ReadFile(filename)
}
