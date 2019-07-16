package logs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type consoleWriter struct {
	consoleWriter *os.File
	level         int
	WriterName    string `json:"writer"`
}

func (w *consoleWriter) WriteMsg(message logMessage) (err error) {
	if message.level < w.level {
		return nil
	}

	var msg string
	if w.level == LevelTrace {
		msg = fmt.Sprintf("%s [%s] [%s] [%s:%d] - %s\n", message.timeString, levelString[message.level],
			message.trace.funcName, message.trace.file, message.trace.line, message.message)
	} else {
		msg = fmt.Sprintf("%v [%v] - %v\n", message.timeString, levelString[message.level], message.message)
	}

	_, err = w.consoleWriter.Write([]byte(msg))
	return
}

func (w *consoleWriter) Flush() {

}

func (w *consoleWriter) Destroy() {

}

func newConsoleAdapter(level string, helper string) (writer *consoleWriter, err error) {

	w := getConsoleWriter()

	if err = json.Unmarshal([]byte(helper), &w); err != nil {
		return
	}

	switch w.WriterName {
	case "stdout":
		w.consoleWriter = os.Stdout
	}

	if w.level = getLevelInt(level); w.level == -1 {
		err = errors.New("not support log record level")
	}

	writer = w
	return
}

func getConsoleWriter() *consoleWriter {
	return &consoleWriter{
		consoleWriter: os.Stderr,
		level:         LevelInfo,
	}
}
