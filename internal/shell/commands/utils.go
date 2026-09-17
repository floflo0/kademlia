package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func getFlags(command *cobra.Command) []string {
	command.InitDefaultHelpFlag()
	flags := make([]string, 0)
	command.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, "--"+flag.Name)
		if flag.Shorthand != "" {
			flags = append(flags, "-"+flag.Shorthand)
		}
	})
	return flags
}
