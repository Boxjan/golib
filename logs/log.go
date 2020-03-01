package logs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	// Only trace level will record the file and line
	LevelTrace int = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
)

const (
	LevelTraceStr   = "trace"
	LevelDebugStr   = "debug"
	LevelInfoStr    = "info"
	LevelWarningStr = "warning"
	LevelErrorStr   = "error"
)

var levelString = []string{"  [trace]", "  [debug]", "   [info]", "[warning]", "  [error]"}

const (
	AdapterFile    = "file"
	AdapterConsole = "console"
)

var NoSupportLevel = errors.New("not support log record level")
var NoSupportAdapter = errors.New("not support adapter")

type traceStruct struct {
	file     string
	line     int
	funcName string
}

type logMessage struct {
	time       time.Time
	timeString string
	level      int
	message    string
	trace      traceStruct
}

type logWriter interface {
	WriteMsg(message logMessage) error
	Flush()
	Destroy()
}

type Logger struct {
	recorder       []logWriter
	recorderCount  int
	logMsgCh       chan logMessage
	logMsgChClosed bool
	wg             sync.WaitGroup
	asyncStart     bool
}

func NewLoggerWithCmdWriter(level string) *Logger {
	logger := NewLogger()
	if err := logger.AddAdapter(AdapterConsole, level, `{}`); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
	}
	return logger
}

func NewLoggerWithCmdWriterWithTraceLevel() *Logger {
	logger := NewLogger()
	if err := logger.AddAdapter(AdapterConsole, LevelTraceStr, `{}`); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
	}
	return logger
}

func NewLogger() *Logger {
	l := &Logger{
		recorderCount:  0,
		asyncStart:     false,
		logMsgChClosed: true,
	}
	runtime.SetFinalizer(l, (*Logger).Close)
	return l
}

func (logger *Logger) AddAdapter(adapterName string, level string, helper string) (err error) {
	if helper == "" {
		helper = `{}`
	}

	var oneWriter logWriter
	switch adapterName {
	case AdapterConsole:
		oneWriter, err = newConsoleAdapter(level, helper)
		break
	case AdapterFile:
		oneWriter, err = newFileAdapter(level, helper)
		break

	default:
		err = NoSupportAdapter

	}
	if err != nil {
		return
	}
	logger.recorder = append(logger.recorder, oneWriter)
	logger.recorderCount++
	return
}

func (logger *Logger) Close() {
	if logger.asyncStart {
		if logger.logMsgChClosed == false {
			close(logger.logMsgCh)
		}
		logger.logMsgChClosed = true
		logger.wg.Wait()
	} else {
		for _, writer := range logger.recorder {
			writer.Destroy()
		}
	}
}

func (logger *Logger) writeMsg(message logMessage) {
	if logger.recorderCount <= 0 {
		_, _ = fmt.Fprint(os.Stderr, "no recorder in the logger\n")
		return
	}

	if logger.asyncStart {
		logger.logMsgCh <- message
	} else {
		for _, writer := range logger.recorder {
			if err := writer.WriteMsg(message); err != nil {
				_, _ = fmt.Fprint(os.Stderr, err)
			}
		}
	}
}

func (logger *Logger) Async() {
	logger.asyncWriteMsg()
}

func (logger *Logger) asyncWriteMsg() {
	if logger.asyncStart == false {
		logger.logMsgCh = make(chan logMessage, 128)
		logger.logMsgChClosed = false
		logger.wg.Add(1)

		go func() {
			for {
				message, ok := <-logger.logMsgCh
				if !ok {
					for _, writer := range logger.recorder {
						writer.Destroy()
					}
					logger.wg.Done()
					return
				}

				for _, writer := range logger.recorder {
					if err := writer.WriteMsg(message); err != nil {
						_, _ = fmt.Fprint(os.Stderr, err)
					}
				}

			}

		}()
		logger.asyncStart = true
	}
}

// use {} as message place
func (logger *Logger) Debug(message string, args ...interface{}) {
	logger.saveLog(LevelDebug, message, args...)
}

func (logger *Logger) Info(message string, args ...interface{}) {
	logger.saveLog(LevelInfo, message, args...)
}

func (logger *Logger) Warning(message string, args ...interface{}) {
	logger.saveLog(LevelWarning, message, args...)
}

func (logger *Logger) Error(message string, args ...interface{}) {
	logger.saveLog(LevelError, message, args...)
}

func (logger *Logger) saveLog(level int, msg string, args ...interface{}) {
	singleLog := logMessage{}
	singleLog.level = level

	singleLog.trace = logTracer()

	singleLog.time = time.Now()
	singleLog.timeString = fmt.Sprintf("%v.%-04d",
		singleLog.time.Format("2006-01-02 15:04:05"),
		singleLog.time.Nanosecond()/100000)

	singleLog.message = parseMessage(msg, args...)

	logger.writeMsg(singleLog)
}

// use format to parse message
func (logger *Logger) DebugF(message string, args ...interface{}) {
	logger.saveLogFormat(LevelDebug, message, args...)
}

func (logger *Logger) InfoF(message string, args ...interface{}) {
	logger.saveLogFormat(LevelInfo, message, args...)
}

func (logger *Logger) WarningF(message string, args ...interface{}) {
	logger.saveLogFormat(LevelWarning, message, args...)
}

func (logger *Logger) ErrorF(message string, args ...interface{}) {
	logger.saveLogFormat(LevelError, message, args...)
}

func (logger *Logger) saveLogFormat(level int, msg string, args ...interface{}) {
	singleLog := logMessage{}
	singleLog.level = level

	singleLog.trace = logTracer()

	singleLog.time = time.Now()
	singleLog.timeString = fmt.Sprintf("%v.%-04d",
		singleLog.time.Format("2006-01-02 15:04:05"),
		singleLog.time.Nanosecond()/100000)

	singleLog.message = fmt.Sprintf(msg, args...)

	logger.writeMsg(singleLog)
}

func parseMessage(message string, args ...interface{}) string {

	sizeOfArgs := len(args)
	sizeOfPlace := strings.Count(message, "{}")

	if sizeOfArgs > sizeOfPlace {
		for i := sizeOfArgs - sizeOfPlace; i > 0; i-- {
			message += " {}"
		}
	} else if sizeOfArgs < sizeOfPlace {
		for i := sizeOfPlace - sizeOfArgs; i > 0; i-- {
			args = append(args, "[Not thing]")
		}
	}

	message = strings.Replace(message, "{}", "%+v", -1)
	return fmt.Sprintf(message, args...)

}

func logTracer() (t traceStruct) {

	var (
		pc uintptr
		ok bool
	)
	pc, t.file, t.line, ok = runtime.Caller(3)

	if ok {
		_, t.file = path.Split(t.file)
		functionNameArray := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		t.funcName = functionNameArray[len(functionNameArray)-1]
	}

	return
}

func getLevelInt(levelStr string) int {
	switch strings.ToLower(levelStr) {
	case LevelTraceStr:
		return LevelTrace
	case LevelDebugStr:
		return LevelDebug
	case LevelInfoStr:
		return LevelInfo
	case LevelWarningStr:
		return LevelWarning
	case LevelErrorStr:
		return LevelError
	}
	return -1
}
