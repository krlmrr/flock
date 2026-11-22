package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

const (
	jsonFlag       = "json"
	jsonAnnotation = "flock.json"
)

func supportsJSON(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[jsonAnnotation] = "true"
	return cmd
}

func jsonOutput(cmd *cobra.Command) bool {
	on, _ := cmd.Flags().GetBool(jsonFlag)
	return on
}

func requireJSONSupport(cmd *cobra.Command, args []string) error {
	if cmd.Annotations[jsonAnnotation] != "" || !jsonRequested(cmd, args) {
		return nil
	}
	return fmt.Errorf("--json isn't supported for `%s`", cmd.CommandPath())
}

func jsonRequested(cmd *cobra.Command, args []string) bool {
	if !cmd.DisableFlagParsing {
		return jsonOutput(cmd)
	}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return false
		}
		if arg == "--"+jsonFlag {
			return true
		}
	}
	return false
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
