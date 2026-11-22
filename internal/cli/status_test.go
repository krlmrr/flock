package cli

import (
	"testing"
	"time"

	"flock/internal/api"
	"flock/internal/config"
	"flock/internal/share"
)

func TestShareLineNamesEveryState(t *testing.T) {
	ends := time.Date(2026, 10, 19, 18, 42, 0, 0, time.Local)
	for _, tc := range []struct {
		st   api.ShareStatus
		want string
	}{
		{api.ShareStatus{Site: "acme", State: share.Building}, "acme: building the assets"},
		{api.ShareStatus{Site: "acme", State: share.Starting}, "acme: starting the tunnel"},
		{api.ShareStatus{Site: "acme", State: share.Failed, Error: "cloudflared exited"}, "acme: failed: cloudflared exited"},
		{api.ShareStatus{Site: "acme", State: share.Running, URL: "https://a-b.trycloudflare.com", ExpiresAt: &ends},
			"acme → https://a-b.trycloudflare.com (ends 6:42PM)"},
		{api.ShareStatus{Site: "acme", State: share.Running, URL: "https://a-b.trycloudflare.com", ExpiresAt: &ends, Allow: []string{"bob@acme.com", "*@client.com"}},
			"acme → https://a-b.trycloudflare.com (ends 6:42PM, only bob@acme.com, *@client.com)"},
	} {
		if got := shareLine(tc.st); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestStatusJSONWhenTheDaemonIsNotRunning(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultPHP, cfg.Xdebug, cfg.Dumps = "8.4", true, false
	flockHome(t, cfg)

	out, err := runFlock(t, "status", "--json")
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "daemon": {
    "running": false,
    "version": null,
    "pid": null,
    "started_at": null
  },
  "default_php": "8.4",
  "xdebug": true,
  "dumps": false,
  "services": []
}
`
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

func TestStatusJSONListsTheInstalledServices(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultPHP = "8.5"
	root := flockHome(t, cfg)
	fakeDaemon(t, root, map[string]any{
		"/v1/status": api.Status{Version: "1.4.0", PID: 4242, StartedAt: time.Date(2026, 11, 22, 9, 30, 0, 0, time.UTC), Xdebug: true, Dumps: true},
		"/v1/services": []api.Service{
			{ID: "web", Name: "Web server (Caddy)", Port: 443, Installed: true, Enabled: true, Running: true},
			{ID: "postgres", Name: "PostgreSQL", Version: "17", Port: 5432, Installed: true, Enabled: true, Stoppable: true, Problem: "exited"},
			{ID: "mysql", Name: "MySQL", Port: 3306, Available: true},
		},
	})

	out, err := runFlock(t, "status", "--json")
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "daemon": {
    "running": true,
    "version": "1.4.0",
    "pid": 4242,
    "started_at": "2026-11-22T09:30:00Z"
  },
  "default_php": "8.5",
  "xdebug": true,
  "dumps": true,
  "services": [
    {
      "id": "web",
      "name": "Web server (Caddy)",
      "version": null,
      "port": 443,
      "enabled": true,
      "running": true,
      "stoppable": false,
      "problem": null
    },
    {
      "id": "postgres",
      "name": "PostgreSQL",
      "version": "17",
      "port": 5432,
      "enabled": true,
      "running": false,
      "stoppable": true,
      "problem": "exited"
    }
  ]
}
`
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}
