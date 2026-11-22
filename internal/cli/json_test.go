package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"flock/internal/api"
	"flock/internal/config"
	"flock/internal/paths"
)

func flockHome(t *testing.T, cfg *config.Config) string {
	t.Helper()
	root, err := os.MkdirTemp("", "flk")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })
	t.Setenv("FLOCK_HOME", root)
	if err := cfg.Save(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func fakeDaemon(t *testing.T, root string, responses map[string]any) {
	t.Helper()
	ln, err := net.Listen("unix", paths.Socket(root))
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	for path, body := range responses {
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, _ *http.Request) {
			json.NewEncoder(w).Encode(body)
		})
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	t.Cleanup(func() { srv.Close() })
}

func runFlock(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestJSONIsRefusedByCommandsWithoutJSONOutput(t *testing.T) {
	flockHome(t, config.Default())
	for _, args := range [][]string{
		{"daemon", "--json"},
		{"logs", "--json"},
		{"share", "--json"},
		{"--json", "share", "acme"},
		{"debug", "--json", "artisan"},
		{"--json", "debug", "artisan"},
		{"version", "--json"},
		{"services", "install", "mysql", "--json"},
	} {
		root := NewRootCommand()
		cmd, _, err := root.Find(args)
		if err != nil {
			t.Fatal(err)
		}
		cmd.RunE = func(*cobra.Command, []string) error {
			t.Errorf("flock %s ran", strings.Join(args, " "))
			return nil
		}
		cmd.Run = nil
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs(args)

		err = root.Execute()
		want := "--json isn't supported for `" + cmd.CommandPath() + "`"
		if err == nil || err.Error() != want {
			t.Errorf("flock %s: err = %v, want %q", strings.Join(args, " "), err, want)
		}
	}
}

func TestDebugPassesJSONAfterTheScriptToPHP(t *testing.T) {
	flockHome(t, config.Default())
	root := NewRootCommand()
	args := []string{"debug", "artisan", "route:list", "--json"}
	cmd, _, err := root.Find(args)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	cmd.RunE = func(_ *cobra.Command, args []string) error {
		got = args
		return nil
	}
	root.SetOut(io.Discard)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if want := []string{"artisan", "route:list", "--json"}; strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("php args = %q, want %q", got, want)
	}
}

func TestServicesJSONKeepsItsShape(t *testing.T) {
	root := flockHome(t, config.Default())
	services := []api.Service{
		{ID: "mysql", Name: "MySQL", Category: "database", Available: true, Version: "8.4.6", Port: 3306,
			Installed: true, Enabled: true, Running: true, Stoppable: true, Env: "DB_CONNECTION=mysql", Log: "mysql", URL: "mysql://root@127.0.0.1:3306"},
		{ID: "redis", Name: "Redis", Category: "cache", Available: false, Unavailable: "not on Windows"},
	}
	fakeDaemon(t, root, map[string]any{"/v1/services": services})

	out, err := runFlock(t, "services", "--json")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.MarshalIndent(services, "", "  ")
	if out != string(want)+"\n" {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}
