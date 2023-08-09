package printer

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gosuri/uilive"
)

type logWriter struct {
	ch          chan []string
	prefix      string
	partialLine string
}

func (w *logWriter) Write(data []byte) (int, error) {
	newLines := strings.Split(string(data), "\n")
	if len(newLines) == 1 {
		w.partialLine = w.partialLine + newLines[0]
		return len(data), nil
	}

	lastLine := newLines[len(newLines)-1]
	readyLines := append([]string{w.partialLine + newLines[0]}, newLines[1:len(newLines)-1]...)
	for i := range readyLines {
		readyLines[i] = w.prefix + readyLines[i]
	}
	w.partialLine = lastLine

	w.ch <- readyLines
	return len(data), nil
}

type lineRingBuffer struct {
	maxSize int
	buffer  []string
}

func (l *lineRingBuffer) push(lines []string) {
	if len(lines) == 0 {
		return
	}

	tmpBuffer := append(l.buffer, lines...)
	if len(tmpBuffer) <= l.maxSize {
		l.buffer = tmpBuffer[:]
	} else {
		l.buffer = tmpBuffer[len(tmpBuffer)-l.maxSize:]
	}
}

func (l *lineRingBuffer) print(w *uilive.Writer) {
	fmt.Fprint(w.Newline(), "\n")
	for _, line := range l.buffer {
		fmt.Fprintf(w.Newline(), "%s\n", line)
	}
}

func (l *lineRingBuffer) reset() {
	l.buffer = []string{}
}

func readAll[T any](ch chan T, fn func(T)) bool {
	for {
		select {
		case element, ok := <-ch:
			if !ok {
				return true
			}
			fn(element)
		default:
			return false
		}
	}
}

type Printer struct {
	actionStdout          *logWriter
	actionStderr          *logWriter
	persistentLogsChannel chan string
	resetChannel          chan bool
}

func New() Printer {
	chActionLogs := make(chan []string, 100)

	result := Printer{
		actionStdout: &logWriter{
			ch:          chActionLogs,
			prefix:      "\033[1;34m[stdout]\033[0m ",
			partialLine: "\033[1;34m[stdout]\033[0m ",
		},
		actionStderr: &logWriter{
			ch:          chActionLogs,
			prefix:      "\033[1;31m[stderr]\033[0m ",
			partialLine: "\033[1;31m[stderr]\033[0m ",
		},
		persistentLogsChannel: make(chan string, 100),
		resetChannel:          make(chan bool, 0),
	}

	go func() {
		writter := uilive.New()
		writter.RefreshInterval = 10 * time.Millisecond
		writter.Start()

		buffer := lineRingBuffer{
			maxSize: 10,
		}
		defer writter.Stop()

		actionChannelReadFn := func(logs []string) {
			buffer.push(logs)
		}

		persistentChannelReadFn := func(line string) {
			fmt.Fprintln(writter.Bypass(), line)
		}

		resetFn := func(ok bool) {
			buffer.reset()
			if err := writter.Flush(); err != nil {
				fmt.Fprintln(writter.Bypass(), err.Error())
			}
		}

		for {
			// TODO: cleanup this logic
			select {
			case data := <-chActionLogs:
				actionChannelReadFn(data)
			case data := <-result.persistentLogsChannel:
				persistentChannelReadFn(data)
			case data := <-result.resetChannel:
				resetFn(data)
			}
			actionFinished := readAll(chActionLogs, actionChannelReadFn)
			persistentFinished := readAll(result.persistentLogsChannel, persistentChannelReadFn)
			readAll(result.resetChannel, resetFn)
			if actionFinished && persistentFinished {
				return
			}

			buffer.print(writter)
			time.Sleep(time.Millisecond * 20)
		}
	}()

	return result
}

func (p *Printer) ActionStdout() io.Writer {
	return p.actionStdout
}

func (p *Printer) ActionStderr() io.Writer {
	return p.actionStderr
}

func (p *Printer) PersistentPrintln(line string) {
	p.persistentLogsChannel <- line
}

func (p *Printer) ResetActionBuffer() {
	p.resetChannel <- true
}
