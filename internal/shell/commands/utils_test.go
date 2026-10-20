package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestGetCompletions(t *testing.T) {
	cmd := &cobra.Command{
		Use: "root",
	}
	cmd.Flags().StringP("flag", "f", "", "test flag")

	subCmd := &cobra.Command{
		Use: "sub",
	}
	subCmd.Flags().Bool("subflag", false, "test sub flag")
	cmd.AddCommand(subCmd)

	items := GetCompletions(cmd)
	if len(items) == 0 {
		t.Fatalf("expected completions, got 0")
	}
}
