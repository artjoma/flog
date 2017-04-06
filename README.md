# flog
Simple and fast asynchronous logging.<br/>
Two types of loggers : Console layout and file layout<br/>
Three log levels: DEBUG, INFO, ERROR<br/>

File logger:<br/>
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
I0604 13:40:03.397765 repair.go:169-Start destroy
I0604 13:40:03.898044 repair.go:173-End destroy
I0604 13:48:27.992297 repair.go:115-Start create services
```
