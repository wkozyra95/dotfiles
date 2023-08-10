package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// NamedLogger creates named package logger.
func NamedLogger(name string) logrus.Logger {
	logLevel := logrus.InfoLevel
	if os.Getenv("DEBUG") != "" {
		logLevel = logrus.DebugLevel
	} else if os.Getenv("TRACE") != "" {
		logLevel = logrus.TraceLevel
	}
	return logrus.Logger{
		Out: os.Stderr,
		Formatter: &CustomTextFormatter{
			logrus.TextFormatter{
				ForceColors: true,
			},
			name,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logLevel,
	}
}

// CustomTextFormatter ...
type CustomTextFormatter struct {
	logrus.TextFormatter
	Name string
}

// Format renders a single log entry
func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level <= logrus.ErrorLevel {
		return []byte(color.RedString(entry.Message + "\n")), nil
	} else if entry.Level <= logrus.WarnLevel {
		return []byte(color.YellowString(entry.Message + "\n")), nil
	} else if entry.Level <= logrus.InfoLevel {
		return []byte(entry.Message + "\n"), nil
	} else {
		_, file, no, _ := runtime.Caller(5)
		entry.Message = fmt.Sprintf("[%-8s][%-15s:%03d]%s", f.Name, path.Base(file), no, entry.Message)
		return f.TextFormatter.Format(entry)
	}
}

type TransparentFormatter struct{}

func (f *TransparentFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%v\n", time.Now()) + entry.Message), nil
}
