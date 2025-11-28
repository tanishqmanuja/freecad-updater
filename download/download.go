package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadWithProgress downloads a file from URL and shows a single-line progress bar
func DownloadWithProgress(url, outPath string, totalSize int64) error {
	tempPath := outPath + ".tmp"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to request URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	outFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	buf := make([]byte, 64*1024)
	var downloaded int64
	barWidth := 40
	start := time.Now()

	printProgress := func() {
		percent := float64(downloaded) / float64(totalSize) * 100
		filled := int(percent / 100 * float64(barWidth))
		empty := barWidth - filled
		bar := fmt.Sprintf("[%s%s]", stringRepeat("█", filled), stringRepeat(" ", empty))
		fmt.Printf("\r %s %.1f%% (%d MB / %d MB)", bar, percent, downloaded/(1024*1024), totalSize/(1024*1024))
	}

	printProgress() // initial

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := outFile.Write(buf[:n]); wErr != nil {
				return fmt.Errorf("write error: %w", wErr)
			}
			downloaded += int64(n)
			printProgress()
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read error: %w", err)
		}
	}

	// Finish line
	duration := time.Since(start).Seconds()
	fmt.Printf("\r [%s] 100.0%% (%d MB / %d MB) - %.1fs\n", stringRepeat("█", barWidth), downloaded/(1024*1024), totalSize/(1024*1024), duration)

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Rename temp file to final
	if err := os.Rename(tempPath, outPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

func stringRepeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
