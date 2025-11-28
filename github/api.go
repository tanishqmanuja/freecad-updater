package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Release struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

type Asset struct {
	BrowserDownloadURL string    `json:"browser_download_url"`
	Size               int64     `json:"size"`
	CreatedAt          time.Time `json:"created_at"`
}

func CheckRateLimit() (int, error) {
	resp, err := http.Get("https://api.github.com/rate_limit")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var rate struct {
		Rate struct {
			Remaining int `json:"remaining"`
		} `json:"rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		return 0, err
	}

	return rate.Rate.Remaining, nil
}

func GetReleases() ([]Release, error) {
	resp, err := http.Get("https://api.github.com/repos/FreeCAD/FreeCAD/releases")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

func GetLatestWeeklyRelease(releases []Release) (*Release, error) {
	var latest *Release
	for _, r := range releases {
		if !strings.HasPrefix(strings.ToLower(r.TagName), "weekly-") {
			continue
		}
		if latest == nil || r.PublishedAt.After(latest.PublishedAt) {
			latest = &r
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no weekly releases found")
	}
	return latest, nil
}

func AssetVersionFromTag(tag string) string {
	parts := strings.Split(tag, "-")
	if len(parts) < 2 {
		return ""
	}
	return strings.ReplaceAll(parts[1], ".", "")
}

func FindWindows7zAsset(release *Release) (*Asset, error) {
	for _, a := range release.Assets {
		lower := strings.ToLower(a.BrowserDownloadURL)
		if strings.Contains(lower, "windows") && strings.HasSuffix(lower, ".7z") {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("no Windows .7z asset found in release %s", release.TagName)
}
