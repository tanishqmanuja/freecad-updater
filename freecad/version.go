package freecad

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// InstalledRevision returns the revision date of the installed FreeCAD executable.
// Example output: "FreeCAD 1.2.0 Revision: 20251126 (Git shallow)"
func InstalledRevision(freecadExe string) (string, error) {
	cmd := exec.Command(freecadExe, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run FreeCAD: %w", err)
	}

	re := regexp.MustCompile(`Revision:\s*(\d{8})`)
	matches := re.FindStringSubmatch(out.String())
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse revision from output: %s", out.String())
	}

	return matches[1], nil
}

// NeedsUpdate compares installed revision with release revision from tag.
// releaseTag is something like "weekly-2025.11.26"
func NeedsUpdate(installedRev string, releaseTag string) bool {
	parts := strings.Split(releaseTag, "-")
	if len(parts) < 2 {
		return true // invalid tag, force update
	}
	rev := strings.ReplaceAll(parts[1], ".", "")
	return installedRev != rev
}
