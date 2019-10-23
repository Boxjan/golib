package logs

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	log := NewLogger()
	var err error

	err = log.AddAdapter("file", LevelTraceStr,
		`{"filename":"log/app.log", "rotate":true, "daily": true, "maxlines": 10, "maxsize": 10240000}`)
	if err != nil {
		t.Error(err)
	}
	err = log.AddAdapter("console", LevelTraceStr, `{}`)
	if err != nil {
		t.Error(err)
	}
	err = log.AddAdapter("console", LevelDebugStr, `{}`)
	if err != nil {
		t.Error(err)
	}

	log.Error("error")
	log.Warning("warning")
	log.Info("informational")
	log.Debug("debug")

	log.Close()
	_ = os.RemoveAll("log/")
}

func TestNewLoggerWithCmdWriter(t *testing.T) {
	log := NewLoggerWithCmdWriter()
	log.Error("error")
	log.Warning("warning")
	log.Info("informational")
	log.Debug("debug")

	log.Close()
}

func BenchmarkGoFormat(b *testing.B) {
	log := NewLogger()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"go-format.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.DebugF("debug %s", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("go-format.log")
}

func BenchmarkGoFormatAsync(b *testing.B) {
	log := NewLogger()
	log.Async()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"go-format-async.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.DebugF("debug %s", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("go-format-async.log")
}

func BenchmarkLoggerFormat(b *testing.B) {
	log := NewLogger()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"logger-format.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.Debug("debug {}", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("logger-format.log")
}

func BenchmarkLoggerFormatAsync(b *testing.B) {
	log := NewLogger()
	log.Async()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"logger-format-async.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.Debug("debug {}", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("logger-format-async.log")
}

func BenchmarkMultipleWriter(b *testing.B) {
	const fileCount = 4

	log := NewLogger()
	for i := 0; i < fileCount; i++ {
		_ = log.AddAdapter("file", "trace", `{"filename":"./bench/bench-`+strconv.Itoa(i)+`.log", "rotate":false}`)
	}

	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}

	log.Close()
	_ = os.RemoveAll("./bench")
}

func BenchmarkMultipleWriterAsync(b *testing.B) {
	const fileCount = 4

	log := NewLogger()
	for i := 0; i < fileCount; i++ {
		_ = log.AddAdapter("file", "trace", `{"filename":"./async-bench/bench-`+strconv.Itoa(i)+`.log", "rotate":false}`)
	}
	log.Async()

	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}

	log.Close()
	_ = os.RemoveAll("./async-bench")
}
