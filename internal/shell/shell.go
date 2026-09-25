package shell

import (
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/logging"
	"kademlia/internal/shell/commands"
	"kademlia/internal/shell/commands/exit"
	"kademlia/internal/shell/commands/ping"
	"kademlia/internal/shell/commands/show"
	"log/slog"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type Shell struct {
	commands map[string]commands.Command
}

func NewShell(kademlia kademlia.Kademlia) *Shell {
	return &Shell{
		commands: map[string]commands.Command{
			"exit": exit.NewExitCommand(),
			"ping": ping.NewPingCommand(kademlia),
			"show": show.NewShowCommand(kademlia, os.Stdout),
		},
	}
}

func toPcItems(
	completions []commands.Completion,
) []readline.PrefixCompleterInterface {
	items := make([]readline.PrefixCompleterInterface, 0, len(completions))
	for _, c := range completions {
		items = append(
			items,
			readline.PcItem(c.Name, toPcItems(c.Children)...),
		)
	}
	return items
}

func (s *Shell) completer() *readline.PrefixCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, len(s.commands))
	for name, command := range s.commands {
		items = append(
			items,
			readline.PcItem(name, toPcItems(command.GetCompletions())...),
		)
	}
	return readline.NewPrefixCompleter(items...)
}

func (s *Shell) Run() error {
	slog.Debug("Start shell")

	lineReader, err := readline.NewEx(&readline.Config{
		Prompt:       "[kademlia] >>> ",
		AutoComplete: s.completer(),
	})
	if err != nil {
		return err
	}
	defer lineReader.Close()
	logging.InitLogger(lineReader.Stdout())
	defer logging.InitLogger(os.Stdout)

	for {
		line, err := lineReader.Readline()
		if err != nil {
			if err == io.EOF {
				break
			}
			if err == readline.ErrInterrupt {
				continue
			}
			// Unreachable
			panic(err)
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		args := strings.Fields(input)
		commandName := args[0]
		command, exists := s.commands[commandName]
		if !exists {
			fmt.Printf("Error: unknown command: %s\n", commandName)
			continue
		}
		command.Execute(args[1:])
	}
	return nil
}
