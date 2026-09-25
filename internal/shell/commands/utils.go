package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func GetCompletions(command *cobra.Command) []Completion {
	command.InitDefaultHelpFlag()
	items := make([]Completion, 0)
	for _, sub := range command.Commands() {
		if sub.IsAvailableCommand() {
			items = append(
				items,
				Completion{
					Name:     sub.Name(),
					Children: GetCompletions(sub),
				},
			)
		}
	}
	command.Flags().VisitAll(func(flag *pflag.Flag) {
		items = append(items, Completion{Name: "--" + flag.Name})
		if flag.Shorthand != "" {
			items = append(items, Completion{Name: "-" + flag.Shorthand})
		}
	})
	return items
}
