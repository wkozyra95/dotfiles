package action

import (
	"bytes"
	"fmt"
	"strings"
)

type actionPrinter struct {
	printFn func(string)
	stack   []string
}

func (p *actionPrinter) startList() {
	p.stack = append(p.stack, "- ")
}

func (p *actionPrinter) endList() {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1] != "- " {
		panic("Unexpected stack state")
	}
	p.stack = p.stack[:len(p.stack)-1]
}

// assume conditions are single line and can be nested
func startSelectNodeCondition[T conditionalKey](p *actionPrinter, node selectNode[T]) {
	p.printFn(
		fmt.Sprintf(
			"%s%s: %s",
			p.prefixFirstLine(), node.selector.conditionName, node.selector.string(),
		),
	)
}

func (p *actionPrinter) startSelectNodeBranch(key string) {
	p.printFn(
		fmt.Sprintf("%s%s:", p.prefixSecondaryLine(), key),
	)
	p.stack = append(p.stack, "  ")
}

func (p *actionPrinter) printSelectNodeNoBranchSelected(key string) {
	p.printFn(
		fmt.Sprintf("%s%s: (nothing to do)", p.prefixSecondaryLine(), key),
	)
}

func (p *actionPrinter) endSelectNodeBranch() {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1] != "  " {
		panic("Unexpected stack state")
	}
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *actionPrinter) startScope(label string) {
	p.printFn(
		fmt.Sprintf("%s%s:", p.prefixFirstLine(), label),
	)
	p.stack = append(p.stack, "  ")
}

func (p *actionPrinter) endScope() {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1] != "  " {
		panic("Unexpected stack state")
	}
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *actionPrinter) printLeafNode(node leafNode) {
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

func (p *actionPrinter) prefixFirstLine() string {
	l := 0
	for _, i := range p.stack[:len(p.stack)-1] {
		l += len(i)
	}
	return strings.Join([]string{
		strings.Repeat(" ", l),
		p.stack[len(p.stack)-1],
	}, "")
}

func (p *actionPrinter) prefixSecondaryLine() string {
	l := 0
	for _, i := range p.stack {
		l += len(i)
	}
	return strings.Repeat(" ", l)
}

func SprintActionTree(o Object) string {
	var buf bytes.Buffer
	printer := &actionPrinter{
		printFn: func(s string) {
			buf.WriteString(s)
			buf.WriteString("\n")
		},
	}

	printer.sprint(o.build())
	return buf.String()
}

func PrintActionTree(o Object) {
	fmt.Print(SprintActionTree(o))
}

func (p *actionPrinter) sprint(n node) {
	switch node := n.(type) {
	case listNode:
		p.startList()
		defer p.endList()
		for _, child := range node.children {
			p.sprint(child)
		}
	case leafNode:
		p.printLeafNode(node)
	case wrappedNode:
		if node.optionalLabel != "" {
			p.startScope(node.optionalLabel)
			defer p.endScope()
		}
		p.sprint(node.child)
	case scopeNode:
		p.startScope(node.label)
		defer p.endScope()
		p.sprint(node.nodeProvider())
	case selectNode[ConditionResultType]:
		startSelectNodeCondition(p, node)
		for selectName, child := range node.children {
			p.startSelectNodeBranch(selectName.String())
			p.sprint(child)
			p.endSelectNodeBranch()
		}
	}
}
