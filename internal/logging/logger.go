package logging

import (
	"io"
	"kademlia/config"
	"log/slog"
	"os"
	"path/filepath"
)

func InitLogger(writer io.Writer) {
	level := new(slog.LevelVar)
	level.Set(config.LogLevel)
	projectRoot, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	replace := func(groups []string, attr slog.Attr) slog.Attr {
		if attr.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		if attr.Key == slog.SourceKey && len(groups) == 0 {
			source := attr.Value.Any().(*slog.Source)
			relativeFilePath, err := filepath.Rel(projectRoot, source.File)
			if err == nil {
				source.File = relativeFilePath
			}
		}
		return attr
	}

	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replace,
	}))
	slog.SetDefault(logger)
}
