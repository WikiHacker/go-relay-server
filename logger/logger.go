package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	logFile *os.File
	logger  *log.Logger
	config  Config
}

type Config struct {
	LogFile  string
	LogLevel LogLevel
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

func NewLogger(logFile string, logLevel LogLevel) (*Logger, error) {
	logger := &Logger{
		config: Config{
			LogFile:  logFile,
			LogLevel: logLevel,
		},
	}

	if err := logger.setupLogger(); err != nil {
		return nil, fmt.Errorf("failed to setup logger: %v", err)
	}

	go logger.DailyLogRotation()

	return logger, nil
}

func (l *Logger) setupLogger() error {
	logFileName := l.getLogFileName()
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	l.logFile = logFile
	l.logger = log.New(logFile, "", log.LstdFlags)

	return nil
}

func (l *Logger) getLogFileName() string {
	currentDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s-%s.log", l.config.LogFile, currentDate)
}

func (l *Logger) DailyLogRotation() {
	for {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		durationUntilMidnight := nextMidnight.Sub(now)

		select {
		case <-time.After(durationUntilMidnight):
			if l.logFile != nil {
				l.logFile.Close()
			}

			if err := l.setupLogger(); err != nil {
				l.Log(LogLevelError, "Error rotating log file: %v", err)
			}

			l.deleteOldLogs(7)
		}
	}
}

func (l *Logger) deleteOldLogs(days int) {
	files, err := os.ReadDir(".")
	if err != nil {
		l.Log(LogLevelError, "Error reading log directory: %v", err)
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasPrefix(file.Name(), l.config.LogFile) && strings.HasSuffix(file.Name(), ".log") {
			fileInfo, err := file.Info()
			if err != nil {
				l.Log(LogLevelError, "Error getting file info for %s: %v", file.Name(), err)
				continue
			}

			if fileInfo.ModTime().Before(cutoffTime) {
				if err := os.Remove(file.Name()); err != nil {
					l.Log(LogLevelError, "Error deleting old log file %s: %v", file.Name(), err)
				} else {
					l.Log(LogLevelInfo, "Deleted old log file: %s", file.Name())
				}
			}
		}
	}
}

func (l *Logger) Log(level LogLevel, format string, args ...interface{}) {
	if l.shouldLog(level) {
		message := fmt.Sprintf(format, args...)
		l.logger.Printf("[%s] %s\n", level, message)
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	switch l.config.LogLevel {
	case LogLevelInfo:
		return true
	case LogLevelWarn:
		return level == LogLevelWarn || level == LogLevelError
	case LogLevelError:
		return level == LogLevelError
	default:
		return true
	}
}