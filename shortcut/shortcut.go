package shortcut

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CreateShortcut creates a Windows shortcut (.lnk) at linkPath
// pointing to targetExe with the specified workingDir.
func CreateShortcut(targetExe, workingDir, linkPath string) error {
	vbs := fmt.Sprintf(`
Set oWS = WScript.CreateObject("WScript.Shell")
Set oLink = oWS.CreateShortcut("%s")
oLink.TargetPath = "%s"
oLink.WorkingDirectory = "%s"
oLink.Save
`, linkPath, targetExe, workingDir)

	// Write temporary VBScript
	temp := filepath.Join(os.TempDir(), "make_link.vbs")
	if err := os.WriteFile(temp, []byte(vbs), 0644); err != nil {
		return fmt.Errorf("failed to write temporary VBScript: %w", err)
	}
	defer os.Remove(temp)

	// Run VBScript via cscript
	cmd := exec.Command("cscript", "//nologo", temp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create shortcut: %w", err)
	}

	return nil
}
