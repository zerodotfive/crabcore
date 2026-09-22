package fetch

import (
	"fmt"
	"io"
	"math"
	"strings"
)

type progressReader struct {
	url       string
	r         io.Reader
	sizeBytes int64
	doneBytes int64
}

func (pr *progressReader) progressPercent(n int) {
	progressBarSize := 50
	percent := int(math.Ceil(float64(pr.doneBytes) * float64(progressBarSize) / float64(pr.sizeBytes)))

	fmt.Printf("\r0%%|%s%s|100%% %s", strings.Repeat("▉", percent), strings.Repeat(" ", progressBarSize-percent), pr.url)

	if n == 0 || pr.doneBytes > pr.sizeBytes {
		fmt.Printf("\n")
	}
}

func (pr *progressReader) progressSinner(n int) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	percent := int(math.Ceil(float64(pr.doneBytes) * 10 / float64(pr.sizeBytes)))

	fmt.Printf("\r%s %s", spinner[percent-1], pr.url)

	if n == 0 || pr.doneBytes > pr.sizeBytes {
		fmt.Printf("\n")
	}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.doneBytes += int64(n)

	if pr.sizeBytes < 0 {
		pr.progressSinner(n)

		return n, err
	}

	pr.progressPercent(n)
	return n, err
}
