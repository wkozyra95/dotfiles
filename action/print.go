package action

import (
	"bytes"
	"fmt"
	"strings"
)

type printer struct {
	printFn func(string)
	stack   []string
}

func (p *printer) startList() {
	p.stack = append(p.stack, "- ")
}

func (p *printer) endList() {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1] != "- " {
		panic("Unexpected stack state")
	}
	p.stack = p.stack[:len(p.stack)-1]
}

// assume conditions are single line and can be nested
func startSelectNodeCondition[T conditionalKey](p *printer, node selectNode[T]) {
	p.printFn(
		fmt.Sprintf(
			"%s%s: %s",
			p.prefixFirstLine(), node.selector.conditionName, node.selector.string(),
		),
	)
}

func (p *printer) startSelectNodeBranch(key string) {
	p.printFn(
		fmt.Sprintf("%s%s:", p.prefixSecondaryLine(), key),
	)
	p.stack = append(p.stack, "  ")
}

func (p *printer) printSelectNodeNoBranchSelected(key string) {
	p.printFn(
		fmt.Sprintf("%s%s: (nothing to do)", p.prefixSecondaryLine(), key),
	)
}

func (p *printer) endSelectNodeBranch() {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1] != "  " {
		panic("Unexpected stack state")
	}
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *printer) startScope(label string) {
	p.printFn(
		fmt.Sprintf("%s%s:", p.prefixFirstLine(), label),
	)
	p.stack = append(p.stack, "  ")
}

func (p *printer) endScope() {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1] != "  " {
		panic("Unexpected stack state")
	}
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *printer) printLeafNode(node leafNode) {
	nodeStrLines := strings.Split(node.String(), "\n")
	firstLine := nodeStrLines[0]
	rest := nodeStrLines[1:]

	p.printFn(
		strings.Join([]string{p.prefixFirstLine(), firstLine}, ""),
	)

	prefix := p.prefixSecondaryLine()
	for _, line := range rest {
		p.printFn(
			strings.Join([]string{prefix, line}, ""),
		)
	}
}

func (p *printer) prefixFirstLine() string {
	l := 0
	for _, i := range p.stack[:len(p.stack)-1] {
		l += len(i)
	}
	return strings.Join([]string{
		strings.Repeat(" ", l),
		p.stack[len(p.stack)-1],
	}, "")
}

func (p *printer) prefixSecondaryLine() string {
	l := 0
	for _, i := range p.stack {
		l += len(i)
	}
	return strings.Repeat(" ", l)
}

func Sprint(o Object) string {
	var buf bytes.Buffer
	printer := &printer{
		printFn: func(s string) {
			buf.WriteString(s)
			buf.WriteString("\n")
		},
	}

	sprint(printer, o.build())
	return buf.String()
}

func Print(o Object) {
	println(Sprint(o))
}

func sprint(printer *printer, n node) {
	switch node := n.(type) {
	case listNode:
		printer.startList()
		defer printer.endList()
		for _, child := range node.children {
			sprint(printer, child)
		}
	case leafNode:
		printer.printLeafNode(node)
	case wrappedNode:
		if node.optionalLabel != "" {
			printer.startScope(node.optionalLabel)
			defer printer.endScope()
		}
		sprint(printer, node.child)
	case scopeNode:
		printer.startScope(node.label)
		defer printer.endScope()
		sprint(printer, node.nodeProvider())
	case selectNode[ConditionResultType]:
		startSelectNodeCondition(printer, node)
		for selectName, child := range node.children {
			printer.startSelectNodeBranch(selectName.String())
			sprint(printer, child)
			printer.endSelectNodeBranch()
		}
	}
}
