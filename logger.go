package flog

/*
	ArtjomA
*/
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"strconv"
)

const (
	LEVEL_DEBUG rune = 'D'
	LEVEL_INFO  rune = 'I'
	LEVEL_ERR   rune = 'E'
)

type LogManager struct {
	logFolder     string
	historyFolder string
	loggers       map[string]*Logger
	maxFileSize   uint64
	consoleLayout bool
}

type Logger struct {
	name         string
	logFile      string
	file         *os.File //open file output stream
	fileSize     uint64
	logManager   *LogManager
	eventChannel chan *LoggerEvent
}
type LoggerEvent struct {
	logger    *Logger
	log       string
	event     rune
	timestamp time.Time
	file      string //caller file
	line      int    //caller line
}

/*
	Create "appHome/log" and "appHome/log/history" folders at app folder.
*/
func NewLogManagerFile(appFolder string, maxFileSize uint64) *LogManager {
	logFolder := filepath.Join(appFolder, "log")
	historyFolder := filepath.Join(logFolder, "history")
	os.MkdirAll(historyFolder, 0777)
	return &LogManager{logFolder, historyFolder, make(map[string]*Logger), maxFileSize, false}
}

func NewLogManagerConsole() *LogManager {
	return &LogManager{"", "", make(map[string]*Logger), 0, true}
}

/*
	Create logger. If file layout, open new file descriptor for this logger
*/
func (self *LogManager) NewLogger(loggerName string) *Logger {
	if logger, ok := self.loggers[loggerName]; ok {
		return logger
	} else {
		var logger *Logger
		eventCh := make(chan *LoggerEvent, 10000)
		if self.consoleLayout {
			logger = &Logger{loggerName, "", nil, 0, self, eventCh}
		} else {
			logFile, file, size := logger.openFile(self.logFolder, loggerName)
			fmt.Println("logger:", logFile, size, "bytes")
			logger = &Logger{loggerName, logFile, file, size, self, eventCh}
		}

		go logger.logWriterTask()
		self.loggers[loggerName] = logger
		return logger
	}
}

const NUMBERS = "0123456789"
func firstZero(i, n int, buf []byte) {
	buf[i+1] = NUMBERS[n%10]
	n /= 10
	buf[i] = NUMBERS[n%10]
}
func fixedDigits(n, i, d int, buf []byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		buf[i+j] = NUMBERS[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		buf[i+j] = '0'
	}
}
func (self *Logger) logWriterTask() {
	channel := self.eventChannel

	var day, hour, minute, second = 0, 0, 0, 0
	var month time.Month = time.January

	for event := range channel {
		_, month, day = event.timestamp.Date()
		hour, minute, second = event.timestamp.Clock()
		sl:=strconv.FormatInt(int64(event.line),10)[:]

		buf := make([]byte, 22, 22 + len(event.file) + len(event.log) + len(sl) + 3)
		buf[0] = byte(event.event)
		firstZero(1, day, buf)
		firstZero(3, int(month), buf)
		buf[5] = byte(' ')
		firstZero(6, hour, buf)
		buf[8] = byte(':')
		firstZero(9, minute, buf)
		buf[11] = byte(':')
		firstZero(12, second, buf)
		buf[14] = byte('.')
		fixedDigits(6, 15, event.timestamp.Nanosecond() / 1000, buf)
		buf[21] = byte(' ')
		c := copy(buf[22:], event.file)
		buf[21+c] = ':'
		buf = append(buf, sl...)
		buf = append(buf, '-')
		sl = event.log[:]
		buf = append(buf, sl...)
		buf = append(buf, '\n')

		if event.logger.logManager.consoleLayout {
			if event.event == LEVEL_ERR{
				os.Stderr.Write(buf)
			} else{
				os.Stdout.Write(buf)
			}
		} else {
			self.logManager.writeToFile(event, buf)
		}
	}
}

func (self *LogManager) writeToFile(event *LoggerEvent, buffer []byte) {
	logger := event.logger
	count, err := logger.file.Write(buffer)
	logger.file.Sync()
	if err == nil {
		logger.fileSize += uint64(count)
		//rotate file
		if logger.fileSize >= self.maxFileSize {
			logger.file.Sync()
			logger.file.Close()
			nowS := strings.Replace(event.timestamp.Format(time.StampMilli), ":", "_", -1)
			newFileName := logger.name + "_" + nowS + ".log"
			tempFileName := filepath.Join(logger.logManager.logFolder, newFileName)
			os.Rename(logger.logFile, tempFileName)
			_, file, size := logger.openFile(self.logFolder, logger.name)
			logger.file = file
			logger.fileSize = size
			//async copy
			go logger.copyFileToHistory(self.logFolder, newFileName, self.historyFolder)
		}
	} else {
		fmt.Println("Err write to file: "+logger.logFile, err)
	}
}

//move temp file to history folder
func (self *Logger) copyFileToHistory(sourcePath string, fileName string, toFolder string) {
	srcFilePath := filepath.Join(sourcePath, fileName)
	fromFile, err := os.Open(srcFilePath)
	if err == nil {
		defer func() {
			fromFile.Close()
			err = os.Remove(srcFilePath)
			if err != nil {
				fmt.Println("[Logger.go copyFileToHistory] err remove:", err, srcFilePath)
			}
		}()

		toFile, err := os.Create(filepath.Join(toFolder, fileName))
		if err == nil {
			defer toFile.Close()
			_, err = io.Copy(toFile, fromFile)
			toFile.Sync()
			if err != nil {
				fmt.Println("[Logger.go copyFileToHistory] Err copy file: ", fileName)
			}
		} else {
			fmt.Println("[Logger.go copyFileToHistory]Err create file: ", toFolder, fileName)
		}
	} else {
		fmt.Println("[Logger.go copyFileToHistory]Err open file: ", sourcePath, fileName)
	}

}

func (self *Logger) GetFileName(file string) string {
	return file[strings.LastIndex(file, "/")+1:]
}

func (self *Logger) Debug(log string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = self.GetFileName(file)
	} else {
		file = "?"
		line = 0
	}

	self.eventChannel <- &LoggerEvent{self, log, LEVEL_DEBUG, time.Now(), file, line}
}

func (self *Logger) Info(log string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = self.GetFileName(file)
	} else {
		file = "?"
		line = 0
	}

	self.eventChannel <- &LoggerEvent{self, log, LEVEL_INFO, time.Now(), file, line}
}

func (self *Logger) Err(log string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = self.GetFileName(file)
	} else {
		file = "?"
		line = 0
	}
	self.eventChannel <- &LoggerEvent{self, log, LEVEL_ERR, time.Now(), file, line}
}

func (self *Logger) openFile(logFolder string, loggerName string) (string, *os.File, uint64) {
	logFile := filepath.Join(logFolder, loggerName+".log")
	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic("Can't create file: " + logFile)
	}
	fileInfo, _ := file.Stat()
	return logFile, file, uint64(fileInfo.Size())
}

func (self *LogManager) DestroyLogManager() {
	for loggerName, logger := range self.loggers {
		close(logger.eventChannel)
		fmt.Println("start close logger: " + loggerName)
		logger.file.Close()
		fmt.Println("end close")
	}
	self.loggers = nil
}

func (self *LogManager) GetLogFolder() string {
	return self.logFolder
}
