# flog
Simple and fast asynchronous logging.
Two types of loggers : Console layout and file layout
Three log levels: DEBUG, INFO, ERROR

File logger:
LoggerA	-channel- FileA<br/>
LoggerB	-channel- FileB<br/>
LoggerN	-channel- FileN<br/>


```go
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

```
Out:
```go
=== RUN   TestConsoleLog
I1112 16:57:30.603387 /home/tjoma/dev/go_workspace/src/github.com/flog/logger_test.go:13-Some text 1 !
E1112 16:57:30.603390 /home/tjoma/dev/go_workspace/src/github.com/flog/logger_test.go:14-Some text 2 !
start close logger: testLoggerA
end close
start close logger: testLoggerB
end close
--- PASS: TestConsoleLog (0.10s)
=== RUN   TestFileLog
logger: /home/tjoma/test/log/testLoggerA.log 110 bytes
logger: /home/tjoma/test/log/testLoggerB.log 110 bytes
start close logger: testLoggerB
end close
start close logger: testLoggerA
end close
--- PASS: TestFileLog (0.10s)
PASS
ok  	github.com/flog	0.202s
Success: process exited with code 0.
```
