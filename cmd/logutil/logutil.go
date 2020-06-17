package logutil

import (
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

// Most of this is stolen shamelessly^W^W adapted from https://github.com/sirupsen/logrus/issues/678#issuecomment-362569561

// WriterHook is a hook that writes logs of specified LogLevels to specified Writer
type WriterHook struct {
	Writer    io.Writer
	LogLevels []log.Level
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *log.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels define on which log levels a hook would trigger
func (hook *WriterHook) Levels() []log.Level {
	return hook.LogLevels
}

// SetLogSplitOutput adds additional hooks to a *logrus.Logger, enabling
// the log stream to be split between stdout and stderr as appropriate
func SetLogSplitOutput(l *log.Logger) {
	// Send all logs to nowhere by default
	l.SetOutput(ioutil.Discard)

	// Send logs with level >=WARN to stderr
	l.AddHook(&WriterHook{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})

	// Send info and debug logs to stdout
	l.AddHook(&WriterHook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
}

// IntToLogLevel returns a log.Level value from an integer-based mapping
func IntToLogLevel(levelInt int) log.Level {
	var toLogLevel = map[int]log.Level{
		0: log.FatalLevel,
		1: log.ErrorLevel,
		2: log.InfoLevel,
		3: log.DebugLevel,
	}

	if level, ok := toLogLevel[levelInt]; ok {
		return level
	}
	//Invalid level specified, default to info
	return log.InfoLevel
}
