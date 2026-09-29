package put

import (
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type putCommand struct {
	kademlia kademlia.Kademlia
	out      io.Writer
}

func NewPutCommand(kademlia kademlia.Kademlia, out io.Writer) commands.Command {
	return &putCommand{
		kademlia: kademlia,
		out:      out,
	}
}

func (c *putCommand) buildCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "put [-h] file_path",
		Short: "Upload the contents of a file under its hash " +
			"as the key.",
		Args:                  cobra.ExactArgs(1),
		Example:               "put file.txt",
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			filePath := args[0]

			slog.Debug("Running put command", "filePath", filePath)

			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			id, err := c.kademlia.Put(string(fileContent))
			if err != nil {
				return err
			}

			fmt.Fprintf(
				c.out,
				"File successfully stored with ID = %q\n",
				id.String(),
			)

			return nil
		},
	}
	command.SetOut(c.out)
	return command
}

func (c *putCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *putCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}
