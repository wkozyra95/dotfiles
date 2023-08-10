package term

import (
	"io"
	"os"
	"syscall"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
)

type DynamicTextView struct {
	opts             DynamicTextViewOptions
	channel          chan logEvent
	persistentStdout persistentInput
	persistentStderr persistentInput
	stdoutPipe       stdioPipe
	stderrPipe       stdioPipe
}

type DynamicTextViewOptions struct {
	MaxLines int

	// Text added before every line of the input, when printing to
	// the terminal
	StdoutPrefix string
	// Length of the prefix defined under Prefix field. This value should
	// not include terminal escape codes used in Prefix
	StdoutPrefixLen int
	// Text added before every line of the input, when printing to
	// the terminal
	StderrPrefix string
	// Length of the prefix defined under Prefix field. This value should
	// not include terminal escape codes used in Prefix
	StderrPrefixLen int
}

func NewDynamicTextView(options DynamicTextViewOptions) (*DynamicTextView, error) {
	channel := make(chan logEvent, 1000)

	stdoutPipe, stdoutErr := newStdioPipe(&dynamicInput{
		channel: channel,
		definition: &dynamicInputDefinition{
			prefix:       options.StdoutPrefix,
			prefixLength: options.StdoutPrefixLen,
		},
	}, os.Stdout)
	if stdoutErr != nil {
		return nil, stdoutErr
	}

	stderrPipe, stderrErr := newStdioPipe(&dynamicInput{
		channel: channel,
		definition: &dynamicInputDefinition{
			prefix:       options.StderrPrefix,
			prefixLength: options.StderrPrefixLen,
		},
	}, os.Stderr)
	if stderrErr != nil {
		return nil, stderrErr
	}

	go runPrinter(channel, printerOptions{
		maxLines: options.MaxLines,
		stderr:   stderrPipe.realStream(),
		stdout:   stdoutPipe.realStream(),
	})

	return &DynamicTextView{
		opts:    options,
		channel: channel,
		persistentStdout: persistentInput{
			channel:  channel,
			isStderr: false,
		},
		persistentStderr: persistentInput{
			channel:  channel,
			isStderr: true,
		},
		stdoutPipe: stdoutPipe,
		stderrPipe: stderrPipe,
	}, nil
}

func (d *DynamicTextView) PersistentStdout() io.Writer {
	return &d.persistentStdout
}

func (d *DynamicTextView) PersistentStderr() io.Writer {
	return &d.persistentStderr
}

func (d *DynamicTextView) Clear() {
	event, handledChannel := newClearEvent()
	d.channel <- event
	<-handledChannel
}

func (d *DynamicTextView) PersistDynamicContent() {
	event, handledChannel := newPersistEvent()
	d.channel <- event
	<-handledChannel
}

func (d *DynamicTextView) CloseView() {
	d.Clear()
	d.stderrPipe.cleanup()
	d.stdoutPipe.cleanup()
	close(d.channel)
}

func getTermWidth() int {
	type windowSize struct {
		rows uint16
		cols uint16
	}
	var size windowSize
	out, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return 0
	}
	defer out.Close()
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, out.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	return int(size.cols)
}

type stdioPipe struct {
	pipeWriter *os.File
	pipeReader *os.File
	stdioPtr   *os.File
	swaped     bool
}

func newStdioPipe(newWriter io.Writer, original *os.File) (stdioPipe, error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return stdioPipe{}, err
	}

	pipe := stdioPipe{
		pipeWriter: pipeWriter,
		pipeReader: pipeReader,
		stdioPtr:   original,
	}
	pipe.swap()
	go func() {
		buf := make([]byte, 1024*16)
		for {
			n, err := pipeReader.Read(buf)
			if err == io.EOF {
				if err := pipeReader.Close(); err != nil {
					panic(spew.Sdump(err))
				}
				return
			}
			if err != nil {
				panic(err)
			}
			if n2, err := newWriter.Write(buf[:n]); err != nil || n2 != n {
				panic(err)
			}
		}
	}()
	return pipe, nil
}

func (s *stdioPipe) realStream() *os.File {
	if s.swaped {
		return s.pipeWriter
	} else {
		return s.stdioPtr
	}
}

func (s *stdioPipe) swap() {
	s.swaped = !s.swaped
	tmp := *s.stdioPtr
	*s.stdioPtr = *s.pipeWriter
	*s.pipeWriter = tmp
}

func (s *stdioPipe) cleanup() {
	if s.swaped {
		s.swap()
	}
	if err := s.pipeWriter.Close(); err != nil {
		panic(err)
	}
}
