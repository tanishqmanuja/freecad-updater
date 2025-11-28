package extract

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// find7z tries to locate 7z.exe on PATH or in common install directories.
func find7z() (string, error) {
	// Try PATH first
	if path, err := exec.LookPath("7z"); err == nil {
		return path, nil
	}

	// Common installation directories
	commonPaths := []string{
		`C:\Program Files\7-Zip\7z.exe`,
		`C:\Program Files (x86)\7-Zip\7z.exe`,
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("7z.exe not found on PATH or in common directories")
}

// Extract7z extracts the given archive to the target directory.
func Extract7z(archive, target string) error {
	fmt.Printf("Extracting %s to %s...\n", archive, target)

	// Ensure target directory exists
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Locate 7z.exe
	sevenZipPath, err := find7z()
	if err != nil {
		return err
	}

	// Run 7z extraction: 7z x archive -oTarget -y
	cmd := exec.Command(sevenZipPath, "x", archive, "-o"+target, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("7z extraction failed: %w", err)
	}

	return nil
}

// MoveExtracted moves the first directory inside 'srcDir' to 'destDir'.
func MoveExtracted(srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil || len(entries) == 0 {
		return fmt.Errorf("no extracted directory found in %s", srcDir)
	}

	root := filepath.Join(srcDir, entries[0].Name())
	os.MkdirAll(destDir, 0755)

	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		dest := filepath.Join(destDir, rel)
		if d.IsDir() {
			os.MkdirAll(dest, 0755)
		} else {
			return os.Rename(p, dest)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Remove leftover
	return os.RemoveAll(root)
}
