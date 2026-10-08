package shell

import (
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/logging"
	"kademlia/internal/shell/commands"
	"kademlia/internal/shell/commands/exit"
	"kademlia/internal/shell/commands/get"
	"kademlia/internal/shell/commands/install"
	"kademlia/internal/shell/commands/ping"
	"kademlia/internal/shell/commands/publish"
	"kademlia/internal/shell/commands/put"
	"kademlia/internal/shell/commands/show"
	"log/slog"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

type Shell struct {
	commands map[string]commands.Command
	kademlia kademlia.Kademlia
}

func NewShell(kademlia kademlia.Kademlia, dns kademlia.DNSVerifier) *Shell {
	return &Shell{
		commands: map[string]commands.Command{
			"exit":    exit.NewExitCommand(kademlia),
			"get":     get.NewGetCommand(kademlia, os.Stdout),
			"install": install.NewInstallCommand(kademlia, os.Stdout),
			"ping":    ping.NewPingCommand(kademlia),
			"publish": publish.NewPublishCommand(kademlia, dns, os.Stdout),
			"put":     put.NewPutCommand(kademlia, os.Stdout),
			"show":    show.NewShowCommand(kademlia, dns, os.Stdout),
		},
		kademlia: kademlia,
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
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		slog.Debug("Stdout is not a tty")
		return nil
	}

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
				err := s.kademlia.Quit()
				if err != nil {
					return err
				}
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
