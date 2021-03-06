package logs

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func testFileCalls(log *Logger) {
	log.Error("error")
	log.Warning("warning")
	log.Info("info")
	log.Debug("debug")
}

func TestSingleFile1(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("file", "trace", `{"filename":"single1.log", "rotate":false}`)

	testFileCalls(log)
	log.Close()

	f, err := os.Open("single1.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = LevelError
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
	_ = f.Close()
	_ = os.Remove("single1.log")

}

func TestSingleFile2(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("file", "info", `{"filename":"single2.log", "rotate":false}`)

	testFileCalls(log)
	log.Close()

	f, err := os.Open("single2.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = LevelError - LevelDebug
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
	_ = f.Close()
	_ = os.Remove("single2.log")
}

func TestDailyRotateFile(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("file", "trace", `{"filename":"./daily/daily.log", "rotate":true, "daily": true}`)

	testFileCalls(log)
	log.Close()

	files, _ := ioutil.ReadDir("./daily/")
	expected := 1
	if len(files) != expected {
		t.Error(len(files), "not "+strconv.Itoa(expected)+" file")
	}

	_ = os.RemoveAll("./daily/")
}

func TestMaxLineRotateFile(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("file", "trace", `{"filename":"./maxlines/maxlines.log", "rotate":true, "maxlines": 1}`)

	testFileCalls(log)
	log.Close()

	files, _ := ioutil.ReadDir("./maxlines/")
	expected := 4
	if len(files) != expected {
		t.Error(len(files), "not "+strconv.Itoa(expected)+" file")
	}

	_ = os.RemoveAll("./maxlines/")
}

func TestMaxSizeRotateFile(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("file", "trace", `{"filename":"./maxsize/maxsize.log", "rotate":true, "maxsize": 10}`)

	testFileCalls(log)
	log.Close()

	files, _ := ioutil.ReadDir("./maxsize/")
	expected := 4
	if len(files) != expected {
		t.Error(len(files), "not "+strconv.Itoa(expected)+" file")
	}

	_ = os.RemoveAll("./maxsize/")
}

func TestAsyncSingleFileWriter(t *testing.T) {
	log := NewLogger()
	_ = log.AddAdapter("file", "trace", `{"filename":"./async.log", "rotate":false}`)
	log.Async()

	testFileCalls(log)
	log.Close()

	f, err := os.Open("./async.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = LevelError
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
	_ = f.Close()

	_ = os.Remove("./async.log")
}

func TestAsyncMultipleFileWriter(t *testing.T) {
	const fileCount = 5
	log := NewLogger()
	for i := 0; i < fileCount; i++ {
		_ = log.AddAdapter("file", "trace", `{"filename":"./async/async`+strconv.Itoa(i)+`.log", "rotate":false}`)
	}
	log.Async()
	testFileCalls(log)
	log.Close()

	for i := 0; i < fileCount; i++ {
		f, err := os.Open("./async/async" + strconv.Itoa(i) + ".log")
		if err != nil {
			t.Fatal(err)
		}
		b := bufio.NewReader(f)
		lineNum := 0
		for {
			line, _, err := b.ReadLine()
			if err != nil {
				break
			}
			if len(line) > 0 {
				lineNum++
			}
		}
		var expected = LevelError
		if lineNum != expected {
			t.Fatal("./async/async"+strconv.Itoa(i)+".log", lineNum, "not "+strconv.Itoa(expected)+" lines")
		}
		_ = f.Close()
	}

	_ = os.RemoveAll("./async/")
}
