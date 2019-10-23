package logs

import (
	"testing"
)

// Try each log level in decreasing order of priority.
func testConsoleCalls(log *Logger) {
	log.Error("error")
	log.Warning("warning")
	log.Info("informational")
	log.Debug("debug")
}

func TestConsole(t *testing.T) {
	log1 := NewLogger()
	_ = log1.AddAdapter("console", LevelTraceStr, "")
	testConsoleCalls(log1)

	log2 := NewLogger()
	_ = log2.AddAdapter("console", LevelInfoStr, "")
	testConsoleCalls(log1)
}

func TestAsyncSingleConsoleWriter(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("console", LevelTraceStr, ``)
	log.Async()

	testConsoleCalls(log)
	log.Close()

}
