package cli

import (
	"os"
	"path/filepath"
	"testing"

	"flock/internal/config"
	"flock/internal/php"
)

func TestPHPListJSONMarksTheDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultPHP = "8.4"
	root := flockHome(t, cfg)
	for _, minor := range []string{"8.3", "8.4"} {
		bin := php.Binary(root, minor)
		if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(bin, nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	out, err := runFlock(t, "php", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "version": "8.4",
    "default": true
  },
  {
    "version": "8.3",
    "default": false
  }
]
`
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

func TestPHPListJSONIsAnEmptyArrayWithoutVersions(t *testing.T) {
	flockHome(t, config.Default())

	out, err := runFlock(t, "php", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if out != "[]\n" {
		t.Errorf("got %q, want %q", out, "[]\n")
	}
}
