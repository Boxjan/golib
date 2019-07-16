package logs

import (
	"os"
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

func BenchmarkLoggerFormat(b *testing.B) {
	log := NewLogger()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"app.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.DebugF("debug %s", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("app.log")
}

func BenchmarkLoggerFormatGo(b *testing.B) {
	log := NewLogger()
	log.Async()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"app.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.DebugF("debug %s", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("app.log")
}

func BenchmarkLogger(b *testing.B) {
	log := NewLogger()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"app.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.Debug("debug {}", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("app.log")
}

func BenchmarkLoggerGo(b *testing.B) {
	log := NewLogger()
	log.Async()
	var err error

	err = log.AddAdapter("file", LevelTraceStr, `{"rotate":false, "filename":"app.log"}`)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		log.Debug("debug {}", time.Now().Format("2006-01-02"))
	}

	log.Close()
	_ = os.RemoveAll("app.log")
}
