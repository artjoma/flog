package flog

import (
	"testing"
	"time"
)

func TestConsoleLog(t *testing.T) {
	logManager := NewLogManagerConsole()
	loggerA := logManager.NewLogger("testLoggerA")
	loggerB := logManager.NewLogger("testLoggerB")

	loggerA.Info("Some text 1 !")
	loggerB.Err("Some text 2 !")

	time.Sleep(100 * time.Millisecond)
	logManager.DestroyLogManager()

}

func TestFileLog(t *testing.T) {
	logManager := NewLogManagerFile("/home/tjoma/test", 1024*1024*5)
	loggerA := logManager.NewLogger("testLoggerA")
	loggerB := logManager.NewLogger("testLoggerB")

	loggerA.Info("loggerA Some text 1 !")
	loggerB.Info("loggerB Some text 2 !")

	time.Sleep(100 * time.Millisecond)
	logManager.DestroyLogManager()
}
