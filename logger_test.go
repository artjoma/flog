package flog

import (
	"testing"
	"time"
)

func TestConsoleLog(t *testing.T) {
	logManager := NewLogManagerConsole()
	logger := logManager.NewLogger("testLogger")

	logger.Info("Some text 1 !")
	logger.Info("Some text 2 !")
	time.Sleep(1 * time.Second)
}
