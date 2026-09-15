package shell

import (
	"fmt"
	"io"
	"kademlia/internal/shell/commands"
	"log/slog"
	"strings"

	"github.com/chzyer/readline"
)

type Shell interface {
	Run() error
}

type shell struct {
	commands map[string]commands.Command
}

func NewShell() Shell {
	return &shell{
		commands: map[string]commands.Command{
			"exit": commands.NewExitCommand(),
		},
	}
}

func (s *shell) completer() *readline.PrefixCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, len(s.commands))
	for name, command := range s.commands {
		flags := command.GetFlags()
		flagItems := make([]readline.PrefixCompleterInterface, 0, len(flags))
		for _, flag := range flags {
			flagItems = append(flagItems, readline.PcItem(flag))
		}
		items = append(items, readline.PcItem(name, flagItems...))
	}
	return readline.NewPrefixCompleter(items...)
}

func (s *shell) Run() error {
	slog.Debug("Start shell")

	lineReader, err := readline.NewEx(&readline.Config{
		Prompt:       "[kademlia] >>> ",
		AutoComplete: s.completer(),
	})
	if err != nil {
		return err
	}
	defer lineReader.Close()

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
