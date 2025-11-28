package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/tanishqmanuja/freecad-updater/download"
	"github.com/tanishqmanuja/freecad-updater/extract"
	"github.com/tanishqmanuja/freecad-updater/freecad"
	"github.com/tanishqmanuja/freecad-updater/github"
	"github.com/tanishqmanuja/freecad-updater/shortcut"
)

func exit(code int) {
	gray := color.New(color.FgHiBlack)
	gray.Println("\nPress enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	os.Exit(code)
}

var Version = "D.E.V"

func main() {
	c := color.New(color.FgCyan, color.Bold)
	c.Println("#=#=#= FreeCAD Updater =#=#=#")
	c.Printf ("#=#------- v %s -------#=#\n", Version)
	os.Exit(0)

	exe, _ := os.Executable()
	scriptPath := filepath.Dir(exe)

	// --- GitHub API rate limit ---
	_, err := github.CheckRateLimit()
	if err != nil {
		color.Red(" Failed: %v", err)
		exit(1)
	}

	// --- Fetch releases ---
	color.Yellow("=> Fetching release list...")
	releases, err := github.GetReleases()
	if err != nil {
		color.Red(" Failed: %v", err)
		exit(1)
	}

	latest, err := github.GetLatestWeeklyRelease(releases)
	if err != nil {
		color.Red("No weekly release found: %v", err)
		exit(1)
	}

	asset, err := github.FindWindows7zAsset(latest)
	if err != nil {
		color.Red("%v", err)
		exit(1)
	}

	gray := color.New(color.FgHiBlack)
	gray.Printf("Size: %d bytes\n", asset.Size)
	gray.Printf("Tag: %s\n", latest.TagName)
	gray.Printf("Asset URL: %s\n", asset.BrowserDownloadURL)

	// --- Check installed revision ---
	freecadCmd := filepath.Join(scriptPath, "internal", "bin", "freecadcmd.exe")
	currentRev, _ := freecad.InstalledRevision(freecadCmd)

	if !freecad.NeedsUpdate(currentRev, latest.TagName) {
		color.Green("FreeCAD is up to date.")
		exit(1)
	}

	if currentRev != "" {
		color.Green("FreeCAD v%s is available. (Installed: %s)", latest.TagName, currentRev)
	} else {
		color.Green("FreeCAD v%s is available.", latest.TagName)
	}
	fmt.Println()

	// --- Download ---
	color.Yellow("=> Downloading FreeCAD...")
	archive := filepath.Join(scriptPath, "freecad_download.7z")

	err = download.DownloadWithProgress(asset.BrowserDownloadURL, archive, asset.Size)
	if err != nil {
		fmt.Println("Download failed:", err)
	}

	color.Green("Download complete.")
	fmt.Println()

	// --- Extract ---
	color.Yellow("=> Extracting archive...")
	unzipDir := filepath.Join(scriptPath, "unzipped")
	os.RemoveAll(unzipDir)
	if err := extract.Extract7z(archive, unzipDir); err != nil {
		color.Red("Extraction failed: %v", err)
		exit(1)
	}
	color.Green("Extraction complete.")
	fmt.Println()

	// --- Move to Internal ---
	internal := filepath.Join(scriptPath, "internal")
	os.RemoveAll(internal)
	if err := extract.MoveExtracted(unzipDir, internal); err != nil {
		color.Red("Failed to move files: %v", err)
		exit(1)
	}

	os.RemoveAll(unzipDir)
	os.Remove(archive)

	// --- Create shortcut ---
	shortcutPath := filepath.Join(scriptPath, "FreeCAD Dev.lnk")
	if err := shortcut.CreateShortcut(filepath.Join(internal, "bin", "freecad.exe"), filepath.Join(internal, "bin"), shortcutPath); err != nil {
		color.Red("Shortcut creation failed: %v", err)
		exit(1)
	}

	color.Green("=#=#=# Update finished #=#=#=")
	time.Sleep(time.Second)
	exit(0)
}
