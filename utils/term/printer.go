package term

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type printer struct {
	persistentBuffer []string
	textBuffer       dynamicTextBuffer
	opts             printerOptions
}

type printerOptions struct {
	maxLines int
	stdout   *os.File
	stderr   *os.File
}

func runPrinter(channel <-chan logEvent, opts printerOptions) {
	p := printer{
		textBuffer: dynamicTextBuffer{
			maxLines:  opts.maxLines,
			termWidth: getTermWidth(),
		},
		opts: opts,
	}
	for {
		shouldClose := p.drainChannel(channel)
		if shouldClose {
			return
		}
		if len(p.persistentBuffer) != 0 || p.textBuffer.hasChanges {
			p.textBuffer.clearDynamicLines(opts.stdout)
		}
		if len(p.persistentBuffer) != 0 {
			fmt.Fprintln(opts.stdout, strings.Join(p.persistentBuffer, "\n"))
			p.persistentBuffer = []string{}
		}
		if len(p.persistentBuffer) != 0 || p.textBuffer.hasChanges {
			p.textBuffer.print(opts.stdout)
		}
		p.textBuffer.hasChanges = false
		time.Sleep(time.Millisecond * 10)
	}
}

func (p *printer) drainChannel(channel <-chan logEvent) (shouldClose bool) {
	event, ok := <-channel
	if !ok {
		return true
	}
	p.handleEvent(event)
	for {
		select {
		case event, ok := <-channel:
			if !ok {
				return true
			}
			p.handleEvent(event)
		default:
			return false
		}
	}
}

func (p *printer) handleEvent(event logEvent) {
	if e, ok := event.(dynamicLogEvent); ok {
		p.textBuffer.push(e)
	} else if e, ok := event.(persistentLogEvent); ok {
		lines := strings.Split(string(e.data), "\n")
		p.persistentBuffer = append(p.persistentBuffer, lines[:len(lines)-1]...)
	} else if e, ok := event.(controlLogEvent); ok {
		if e.isClearEvent() {
			p.textBuffer.clearDynamicLines(p.opts.stdout)
			e.responseChannel <- struct{}{}
		} else if e.isPersistEvent() {
			// make sure buffer is printed
			p.textBuffer.clearDynamicLines(p.opts.stdout)
			p.textBuffer.print(p.opts.stdout)

			p.textBuffer.writtenLinesCount = 0
			// after setting writtenLinesCount to zero
			// this will just clear current line
			p.textBuffer.clearDynamicLines(p.opts.stdout)
			fmt.Fprintln(p.opts.stdout, "")
			e.responseChannel <- struct{}{}
		} else {
			panic("unknown control event")
		}
	} else {
		panic("unknown event")
	}
}

type dynamicLogLine struct {
	line       string
	definition *dynamicInputDefinition
}

type dynamicTextBuffer struct {
	lines             []dynamicLogLine
	writtenLinesCount int
	maxLines          int
	termWidth         int
	hasChanges        bool
}

func (d *dynamicTextBuffer) push(event dynamicLogEvent) {
	if len(event.data) == 0 {
		return
	}
	d.hasChanges = true
	rawLines := strings.Split(string(event.data), "\n")

	if len(d.lines) > 1 && d.lines[len(d.lines)-1].definition == event.input {
		// combine last line of previous batch with first line of the new
		d.lines[len(d.lines)-1].line = d.lines[len(d.lines)-1].line + rawLines[0]
	} else {
		d.lines = append(d.lines, dynamicLogLine{line: rawLines[0], definition: event.input})
	}

	for _, l := range rawLines[1:] {
		d.lines = append(d.lines, dynamicLogLine{line: l, definition: event.input})
	}
	d.lines = d.lines[max(0, len(d.lines)-d.maxLines):]
}

func (d *dynamicTextBuffer) clearDynamicLines(stdout io.Writer) {
	if d.writtenLinesCount == 0 {
		stdout.Write([]byte("\x1B[0G\x1B[0J"))
		return
	}
	stdout.Write([]byte(fmt.Sprintf("\x1B[%dA\x1B[0G\x1B[0J", d.writtenLinesCount-1)))
	d.writtenLinesCount = 0
}

func (d *dynamicTextBuffer) print(stdout io.Writer) {
	linesToWrite := d.maxLines - 1
	bufferToPrint := []string{}

	for _, logLine := range d.lines {
		segmentLength := d.termWidth - logLine.definition.prefixLength
		lineSegmentsCount := 1 + (len(logLine.line)-1)/segmentLength
		for i := 0; i < lineSegmentsCount; i++ {
			startIndex := i * segmentLength
			endIndex := min(len(logLine.line), (i+1)*segmentLength)
			nextLine := strings.Join([]string{
				logLine.definition.prefix,
				logLine.line[startIndex:endIndex],
			}, "")
			bufferToPrint = append(bufferToPrint, nextLine)
		}
	}

	bufferToPrint = bufferToPrint[max(0, len(bufferToPrint)-linesToWrite):]
	stdout.Write([]byte("\n" + strings.Join(bufferToPrint, "\n")))
	d.writtenLinesCount = len(bufferToPrint) + 1
}
