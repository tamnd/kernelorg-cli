package kernelorg_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/kernelorg-cli/kernelorg"
)

const fixtureJSON = `{
  "latest_stable": {"version": "6.9.7"},
  "releases": [
    {
      "moniker": "stable",
      "version": "6.9.7",
      "released": {"isodate": "2024-06-21", "timestamp": "1718928000"},
      "source": "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.9.7.tar.xz",
      "pgp": "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.9.7.tar.sign",
      "changelog": "https://cdn.kernel.org/pub/linux/kernel/v6.x/ChangeLog-6.9.7",
      "gitweb": "https://git.kernel.org/torvalds/t/linux-6.9.7.tar.gz",
      "eos": ""
    },
    {
      "moniker": "mainline",
      "version": "6.10-rc5",
      "released": {"isodate": "2024-06-23", "timestamp": "1719100800"},
      "source": "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.10-rc5.tar.xz",
      "pgp": "",
      "changelog": "",
      "gitweb": "https://git.kernel.org/torvalds/t/linux-6.10-rc5.tar.gz",
      "eos": ""
    },
    {
      "moniker": "longterm",
      "version": "6.1.97",
      "released": {"isodate": "2024-06-21", "timestamp": "1718928000"},
      "source": "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.97.tar.xz",
      "pgp": "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.97.tar.sign",
      "changelog": "https://cdn.kernel.org/pub/linux/kernel/v6.x/ChangeLog-6.1.97",
      "gitweb": "https://git.kernel.org/stable/t/linux-6.1.97.tar.gz",
      "eos": "2026-12"
    },
    {
      "moniker": "longterm",
      "version": "5.15.162",
      "released": {"isodate": "2024-06-21", "timestamp": "1718928000"},
      "source": "https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.15.162.tar.xz",
      "pgp": "https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.15.162.tar.sign",
      "changelog": "https://cdn.kernel.org/pub/linux/kernel/v5.x/ChangeLog-5.15.162",
      "gitweb": "https://git.kernel.org/stable/t/linux-5.15.162.tar.gz",
      "eos": "2026-10"
    }
  ]
}`

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixtureJSON))
	}))
}

func newTestClient(ts *httptest.Server) *kernelorg.Client {
	cfg := kernelorg.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return kernelorg.NewClient(cfg)
}

func TestReleases_all(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := newTestClient(ts)
	releases, err := c.Releases(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) < 3 {
		t.Errorf("expected >= 3 releases, got %d", len(releases))
	}
}

func TestReleases_stable(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := newTestClient(ts)
	releases, err := c.Releases(context.Background(), "stable")
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) == 0 {
		t.Fatal("expected at least one stable release")
	}
	for _, r := range releases {
		if r.Moniker != "stable" {
			t.Errorf("release %s has moniker %q, want stable", r.Version, r.Moniker)
		}
	}
}

func TestReleases_longterm(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := newTestClient(ts)
	releases, err := c.Releases(context.Background(), "longterm")
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) == 0 {
		t.Fatal("expected at least one longterm release")
	}
	for _, r := range releases {
		if r.Moniker != "longterm" {
			t.Errorf("release %s has moniker %q, want longterm", r.Version, r.Moniker)
		}
		if r.EOL == "" {
			t.Errorf("longterm release %s has empty EOL", r.Version)
		}
	}
}

func TestGetRelease_found(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := newTestClient(ts)
	r, err := c.GetRelease(context.Background(), "6.1.97")
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != "6.1.97" {
		t.Errorf("Version = %q, want 6.1.97", r.Version)
	}
	if r.Moniker != "longterm" {
		t.Errorf("Moniker = %q, want longterm", r.Moniker)
	}
}

func TestGetRelease_notfound(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.GetRelease(context.Background(), "0.0.0")
	if !errors.Is(err, kernelorg.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLatestStable(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := newTestClient(ts)
	ls, err := c.LatestStable(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ls.Version != "6.9.7" {
		t.Errorf("Version = %q, want 6.9.7", ls.Version)
	}
}
