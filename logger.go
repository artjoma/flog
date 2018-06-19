package flog

/*
	ArtjomAminov Fast async log
*/
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	LEVEL_DEBUG rune = 'D' //68
	LEVEL_INFO  rune = 'I' //73
	LEVEL_ERR   rune = 'E' //69

	NUMBERS = "0123456789"
)

type LogManager struct {
	logFolder     string
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
	treshold     rune
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
	os.Mkdir(logFolder, os.ModePerm)
	return &LogManager{logFolder, make(map[string]*Logger), maxFileSize, false}
}

func NewLogManagerConsole() *LogManager {
	return &LogManager{"", make(map[string]*Logger), 0, true}
}

/*
	Create logger. If file layout, open new file descriptor for this logger
*/
func (self *LogManager) NewLogger(loggerName string, threshold rune) *Logger {
	if logger, ok := self.loggers[loggerName]; ok {
		return logger
	} else {
		var logger *Logger
		eventCh := make(chan *LoggerEvent, 10000)
		if self.consoleLayout {
			logger = &Logger{loggerName, "", nil, 0, self, eventCh, threshold}
		} else {
			logFile, file, size := logger.openFile(self.logFolder, loggerName)
			fmt.Println("logger:", logFile, size, "bytes")
			logger = &Logger{loggerName, logFile, file, size, self, eventCh, threshold}
		}

		go logger.logWriterTask()
		self.loggers[loggerName] = logger
		return logger
	}
}

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
	var lineNumberSl string
	var buf []byte

	for event := range channel {
		_, month, day = event.timestamp.Date()
		hour, minute, second = event.timestamp.Clock()
		lineNumberSl = strconv.FormatInt(int64(event.line), 10)

		buf = make([]byte, 22, 22+len(event.file)+len(event.log)+len(lineNumberSl)+3)
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
		fixedDigits(6, 15, event.timestamp.Nanosecond()/1000, buf)
		buf[21] = byte(' ')
		sl := event.file
		buf = append(buf, sl...)
		buf = append(buf, ':')
		buf = append(buf, lineNumberSl...)
		buf = append(buf, '-')
		sl = event.log
		buf = append(buf, sl...)
		buf = append(buf, '\n')

		if event.logger.logManager.consoleLayout {
			if event.event == LEVEL_ERR {
				os.Stderr.Write(buf)
			} else {
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
		}
	} else {
		fmt.Println("Err write to file: "+logger.logFile, err)
	}
}

func (self *Logger) GetFileName(file string) string {
	return file[strings.LastIndex(file, "/")+1:]
}

func (self *Logger) Debug(log string) {
	if self.treshold == LEVEL_DEBUG {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			file = self.GetFileName(file)
		} else {
			file = "?"
			line = 0
		}

		self.eventChannel <- &LoggerEvent{self, log, LEVEL_DEBUG, time.Now(), file, line}
	}
}

func (self *Logger) Info(log string) {
	if self.treshold == LEVEL_DEBUG || self.treshold == LEVEL_INFO {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			file = self.GetFileName(file)
		} else {
			file = "?"
			line = 0
		}
		self.eventChannel <- &LoggerEvent{self, log, LEVEL_INFO, time.Now(), file, line}
	}
}

func (self *Logger) InfoReqId(requestId string, log string){
	if self.treshold == LEVEL_DEBUG || self.treshold == LEVEL_INFO {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			file = self.GetFileName(file)
		} else {
			file = "?"
			line = 0
		}
		self.eventChannel <- &LoggerEvent{self, "[" + requestId + "] " + log, LEVEL_INFO, time.Now(), file, line}
	}
}

func (self *Logger) InfoS(params ...interface{}) {
	if self.treshold == LEVEL_DEBUG || self.treshold == LEVEL_INFO {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			file = self.GetFileName(file)
		} else {
			file = "?"
			line = 0
		}
		self.eventChannel <- &LoggerEvent{self, fmt.Sprint(params...), LEVEL_INFO, time.Now(), file, line}
	}
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

func (self *Logger) ErrReqId(requestId string, log string){
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = self.GetFileName(file)
	} else {
		file = "?"
		line = 0
	}
	self.eventChannel <- &LoggerEvent{self, "[" + requestId + "] " + log, LEVEL_ERR, time.Now(), file, line}
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
