package action

import (
	"fmt"
	"strings"
)

type printer struct {
	printFn                     func(string)
	depth                       int
	current                     int
	shouldAddNewlineIfMultiline bool
	isMultiline                 bool
}

func (p *printer) startLevel() {
	p.depth += 2
}

func (p *printer) startMultilineLevel() {
	p.depth += 2
	p.isMultiline = true
}

func (p *printer) addNewlineIfMultiline() {
	p.shouldAddNewlineIfMultiline = true
}

func (p *printer) endLevel() {
	p.depth -= 2
}

func (p *printer) print(textArg string) {
	text := textArg
	if p.shouldAddNewlineIfMultiline {
		if p.isMultiline || strings.Contains(textArg, "\n") {
			text = "\n" + text
		}
		p.shouldAddNewlineIfMultiline = false
	}
	p.isMultiline = false
	split := strings.Split(text, "\n")
	if split[0] != "" {
		suffixLength := p.depth - p.current
		if suffixLength < 0 {
			suffixLength = 0
		}
		split[0] = fmt.Sprintf("%s%s", strings.Repeat(" ", suffixLength), split[0])
	}
	for i := range split[1:] {
		split[i+1] = fmt.Sprintf("%s%s", strings.Repeat(" ", p.depth), split[i+1])
	}
	result := strings.Join(split, "\n")

	if len(split) == 1 {
		p.current += len(split[0])
	} else {
		p.current = len(split[len(split)-1])
	}
	p.printFn(result)
}
