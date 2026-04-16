package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

const (
	Intent    uint8 = iota // Before entering an operation
	Info                   // After escaping an operation with a success
	Warn                   // After escaping an operation with an ignorable warning
	Error                  // After escaping an operation with an error
	Benchmark              // After an operation for benchmarking
	Test                   // During an operation for debugging internally
	Blocked                // When an operation has been blocked without any error
)

type Logger struct {
	Name    string
	channel chan LogEntry
	file    *os.File
	encoder sonic.Encoder
	date    time.Time
}

type LogEntry struct {
	Time       time.Time `json:"-"`
	Content    string    `json:"c"`
	Level      uint8     `json:"l"`
	Identifier string    `json:"i"`
	FileName   string    `json:"f"`
	TimeString string    `json:"t"`
}

const (
	MaxAge         = 7 * 24 * time.Hour
	FilenameFormat = "2006 01 02"
	TimeFormat     = "150405.0000"
	Path           = "./logs"
)

func CreateLogger(name string) *Logger {
	_logger := &Logger{
		Name:    name,
		channel: make(chan LogEntry, 128),
	}
	_logger.reset()
	go _logger.listen()
	return _logger
}

var RootLogger = CreateLogger("root")

func (logger *Logger) enforceMaxAge() {
	files, err := os.ReadDir(Path)
	if err != nil {
		return
	}

	cutoff := time.Now().UTC().Add(-MaxAge)

	for _, f := range files {
		if f.IsDir() {
			continue // skip folders
		}

		name := f.Name()

		if !strings.HasPrefix(name, logger.Name) {
			continue // skip other logger files
		}

		var startTime time.Time

		startTime, err = time.Parse(FilenameFormat, strings.TrimPrefix(name, logger.Name+" "))
		if err != nil {
			continue // skip unknown files
		}

		if startTime.Before(cutoff) {
			_ = os.Remove(filepath.Join(Path, name)) // delete older files
		}
	}
}

func (logger *Logger) reset() error {
	if logger.file != nil {
		_ = logger.file.Close()
	}
	logger.date = time.Now().UTC().Truncate(24 * time.Hour)
	logger.enforceMaxAge()
	_ = os.Mkdir(Path, 0744)
	fileName := filepath.Join(Path, logger.Name+" "+logger.date.Format(FilenameFormat))
	var err error
	logger.file, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("LOGGER: open file error", err.Error())
		return err
	}
	_, _ = logger.file.WriteString("# ----------------\n# LOGGER START\n# ----------------\n")
	logger.encoder = sonic.ConfigFastest.NewEncoder(logger.file)
	return nil
}

func (logger *Logger) listen() {
	for entry := range logger.channel {
		yl, ml, dl := logger.date.Date()
		ye, me, de := entry.Time.Date()
		if dl != de || ml != me || yl != ye {
			for {
				err := logger.reset()
				if err != nil {
					fmt.Println("LOGGER: reset error", err.Error())
					time.Sleep(time.Second)
					continue
				}
				break
			}
		}
		fmt.Println(levelToString(entry.Level) +
			" [" + entry.Time.Format(time.DateTime) + "] " +
			entry.FileName + " " +
			entry.Identifier + " " +
			entry.Content)
		err := logger.encoder.Encode(entry)
		if err != nil {
			fmt.Println("LOGGER: encode error", err.Error())
			continue
		}
	}
}

func (logger *Logger) Add(level uint8, fileName string, identifier string, content string) LogEntry {
	now := time.Now().UTC()
	entry := LogEntry{
		Time:       now,
		TimeString: now.Format(TimeFormat),
		Level:      level,
		FileName:   fileName,
		Identifier: identifier,
		Content:    content,
	}
	select {
	case logger.channel <- entry:
	default:
		logger.channel <- entry
	}
	return entry
}

func levelToString(level uint8) string {
	switch level {
	case Intent:
		return "INTENT"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Benchmark:
		return "BENCHMARK"
	case Test:
		return "TEST"
	case Blocked:
		return "BLOCKED"
	default:
		return "UNKNOWN"
	}
}
