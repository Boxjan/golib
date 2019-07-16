package logs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type fileWriter struct {

	// lock
	sync.RWMutex

	//different adapter will have different record level
	level int

	// The opened file
	filenameOnly  string
	fileExt       string
	writeFileName string
	fileWriter    *os.File
	Filename      string `json:"filename"`

	// Rotate at line
	MaxLines         int `json:"maxlines"`
	maxLinesCurLines int

	// Rotate at size
	MaxSize        int `json:"maxsize"`
	maxSizeCurSize int

	// Rotate daily
	Daily         bool `json:"daily"`
	dailyString   string
	dailyOpenTime time.Time

	Rotate bool `json:"rotate"`
}

func (w *fileWriter) WriteMsg(message logMessage) (err error) {
	if message.level < w.level {
		return nil
	}

	if w.needRotate() {
		w.Lock()
		err = w.doRotate()
		w.Unlock()
	}

	var msg string
	if w.level == LevelTrace {
		msg = fmt.Sprintf("%s [%s] [%s] [%s:%d] - %s\n", message.timeString, levelString[message.level],
			message.trace.funcName, message.trace.file, message.trace.line, message.message)
	} else {
		msg = fmt.Sprintf("%v [%v] - %v\n", message.timeString, levelString[message.level], message.message)
	}

	w.Lock()
	_, err = w.fileWriter.Write([]byte(msg))
	if err == nil {
		w.maxLinesCurLines++
		w.maxSizeCurSize += len(msg)
	}
	w.Unlock()

	return

}

func (w *fileWriter) Flush() {
	_ = w.fileWriter.Sync()
}

func (w *fileWriter) Destroy() {
	_ = w.fileWriter.Close()
}

func newFileAdapter(level string, helper string) (writer *fileWriter, err error) {
	w := getFileWrite()

	if err = json.Unmarshal([]byte(helper), &w); err != nil {
		return
	}
	w.fileExt = filepath.Ext(w.Filename)
	w.filenameOnly = strings.TrimSuffix(w.Filename, w.fileExt)

	if w.level = getLevelInt(level); w.level == -1 {
		err = errors.New("not support log record level")
		return
	}

	err = w.startLog()
	writer = w
	return
}

func (w *fileWriter) startLog() (err error) {

	if err = os.MkdirAll(filepath.Dir(w.Filename), os.FileMode(0755)); err != nil {
		return
	}

	if w.fileWriter, err = w.OpenFile(); err != nil {
		return
	}

	return w.initFd()
}

func (w *fileWriter) initFd() error {

	fd := w.fileWriter
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s\n", err)
	}

	w.maxSizeCurSize = int(fInfo.Size())
	w.maxLinesCurLines = 0

	if fInfo.Size() > 0 {
		count, err := w.lines()
		if err != nil {
			return err
		}
		w.maxLinesCurLines = count
	}

	if w.needRotate() {
		return w.doRotate()
	}
	return nil

}

func (w *fileWriter) OpenFile() (*os.File, error) {

	w.dailyOpenTime = time.Now()
	w.dailyString = w.dailyOpenTime.Format("2006-01-02")
	w.writeFileName = w.getWriteFileName()

	fd, err := os.OpenFile(w.writeFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0664))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		_ = os.Chmod(w.writeFileName, os.FileMode(0664))
	}

	return fd, err
}

func (w *fileWriter) lines() (int, error) {
	fd, err := os.Open(w.writeFileName)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

func (w *fileWriter) doRotate() (err error) {

	_ = w.fileWriter.Close()
	// use date-timestamp to rename old file
	var fName string
	if time.Now().Second() == w.dailyOpenTime.Second() {
		fName = w.filenameOnly + "-" + w.dailyString + "-" +
			strconv.FormatInt(w.dailyOpenTime.Unix(), 10) + "-" +
			strconv.FormatInt(int64(w.dailyOpenTime.Nanosecond()), 10) + w.fileExt

	} else {
		fName = w.filenameOnly + "-" + w.dailyString + "-" + strconv.FormatInt(w.dailyOpenTime.Unix(), 10) + w.fileExt
	}

	err = os.Rename(w.writeFileName, fName)

	if err != nil {
		goto RESTART_LOG
	}
	err = os.Chmod(fName, os.FileMode(0444))

RESTART_LOG:
	w.startLog()

	return
}

func (w *fileWriter) needRotate() bool {

	return w.Rotate && ((w.Daily && w.dailyString != time.Now().Format("2006-01-02")) ||
		(w.MaxSize > 0 && w.maxSizeCurSize >= w.MaxSize) ||
		(w.MaxLines > 0 && w.maxLinesCurLines >= w.MaxLines))
}

func getFileWrite() *fileWriter {
	return &fileWriter{
		Daily:    true,
		Filename: "app.log",
		Rotate:   true,
		level:    LevelInfo,
	}
}

func (w fileWriter) getWriteFileName() string {
	if w.Rotate {
		return w.filenameOnly + "-" + w.dailyString + w.fileExt
	} else {
		return w.Filename
	}
}
