package logging

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestInitLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	InitLogger(buf)
	slog.Info("test message")
	if buf.Len() == 0 {
		t.Fatalf("expected log output, got empty buffer")
	}
}
