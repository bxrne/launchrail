package logs

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

type LevelType int

func (l LevelType) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
)

type Logger struct {
	name    string
	level   int
	file    *os.File
	stdout  *log.Logger
	fileout *log.Logger
}

// NOTE: Logger entry point
func NewLogger(name string, level int) (*Logger, error) {
	if _, err := os.Stat(".logs"); os.IsNotExist(err) {
		os.Mkdir(".logs", 0755)
	}

	file, err := os.Create(fmt.Sprintf(".logs/%s.log", name))
	if err != nil {
		return nil, err
	}

	return &Logger{
		name:    name,
		level:   level,
		file:    file,
		stdout:  log.New(os.Stdout, "", log.LstdFlags),
		fileout: log.New(file, "", log.LstdFlags),
	}, nil
}

func (l *Logger) GetName() string {
	return l.name
}

func (l *Logger) GetLevel() int {
	return l.level
}

func (l *Logger) SetOutput(out *os.File) {
	l.fileout.SetOutput(out)
}

func (l *Logger) SetBufferOutput(buf *bytes.Buffer) {
	l.fileout.SetOutput(buf)
}

func (l *Logger) SetPrefix(prefix string) {
	l.stdout.SetPrefix(prefix)
	l.fileout.SetPrefix(prefix)
}

func (l *Logger) SetFlags(flags int) {
	l.stdout.SetFlags(flags)
	l.fileout.SetFlags(flags)
}

// NOTE: Log colors to stdout and plaintext to file
func (l *Logger) logMessage(level int, color string, msg string) {
	if level >= l.level {
		consoleMsg := fmt.Sprintf(" | %s | %s%s%s | %s%s%s", l.name, color, LevelType(level), Reset, color, msg, Reset)
		formattedMsg := fmt.Sprintf(" | %s | %s | %s", LevelType(level), l.name, msg)

		l.stdout.Println(consoleMsg)
		l.fileout.Println(formattedMsg)
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.logMessage(INFO, Green, fmt.Sprint(v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.logMessage(WARN, Yellow, fmt.Sprint(v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.logMessage(ERROR, Red, fmt.Sprint(v...))
}

func (l *Logger) Debug(v ...interface{}) {
	l.logMessage(DEBUG, Blue, fmt.Sprint(v...))
}

func (l *Logger) Close() {
	l.file.Close()
}
