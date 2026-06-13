// Package kernelorg provides a client for the kernel.org releases JSON API.
package kernelorg

import "errors"

// ErrNotFound is returned by GetRelease when the version does not exist.
var ErrNotFound = errors.New("kernel: release not found")

// Release holds metadata for a single Linux kernel release.
type Release struct {
	Version      string `json:"version"`
	Moniker      string `json:"moniker"`
	ReleasedDate string `json:"released"`
	SourceURL    string `json:"source_url"`
	PGPSignURL   string `json:"pgp_url"`
	ChangelogURL string `json:"changelog_url"`
	GitwebURL    string `json:"gitweb_url"`
	EOL          string `json:"eol"`
}

// LatestStable holds the latest stable kernel version.
type LatestStable struct {
	Version string `json:"version"`
}

// wireResponse is the top-level JSON document from /releases.json.
type wireResponse struct {
	LatestStable wireLatest    `json:"latest_stable"`
	Releases     []wireRelease `json:"releases"`
}

type wireLatest struct {
	Version string `json:"version"`
}

type wireRelease struct {
	Moniker   string       `json:"moniker"`
	Version   string       `json:"version"`
	Released  wireReleased `json:"released"`
	Source    string       `json:"source"`
	PGP       string       `json:"pgp"`
	Changelog string       `json:"changelog"`
	DiffView  string       `json:"diffview"`
	Gitweb    string       `json:"gitweb"`
	EOL       string       `json:"eos"`
}

type wireReleased struct {
	Isodate   string `json:"isodate"`
	Timestamp int64 `json:"timestamp"`
}

func toRelease(w wireRelease) Release {
	return Release{
		Version:      w.Version,
		Moniker:      w.Moniker,
		ReleasedDate: w.Released.Isodate,
		SourceURL:    w.Source,
		PGPSignURL:   w.PGP,
		ChangelogURL: w.Changelog,
		GitwebURL:    w.Gitweb,
		EOL:          w.EOL,
	}
}
