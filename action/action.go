package action

import (
	"io"

	"github.com/wkozyra95/dotfiles/action/printer"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var log = logger.NamedLogger("action")

type Context struct {
	Stdout io.Writer
	Stderr io.Writer
}

type internalCtx struct {
	printer   *actionPrinter
	publicCtx Context
}

type Condition interface {
	check(internalCtx) (bool, error)
	string() string
}

type Object interface {
	build() node
}

type List []Object

func (l List) build() node {
	children := []node{}
	for _, child := range l {
		if listChild, isList := child.(List); isList {
			children = append(children, listChild.build().(listNode).children...)
		} else {
			children = append(children, child.build())
		}
	}
	return listNode{children}
}

type optional struct {
	object Object
	label  string
}

func Optional(label string, o Object) Object {
	return optional{
		object: o,
		label:  label,
	}
}

func (o optional) build() node {
	return wrappedNode{
		child:         o.object.build(),
		optionalLabel: o.label,
		wrapper: func(ctx internalCtx, innerNode node) error {
			err := innerNode.run(ctx)
			if err != nil {
				log.Error(err)
				if !prompt.ConfirmPrompt("Action failed, do you want to continue?") {
					return err
				}
			}
			return nil
		},
	}
}

type WithCondition struct {
	If   Condition
	Then Object
	Else Object
}

type ConditionResultType bool

func (c ConditionResultType) String() string {
	if c {
		return "Then"
	} else {
		return "Else"
	}
}

func (a WithCondition) build() node {
	children := map[ConditionResultType]node{}
	if a.Then != nil {
		children[true] = a.Then.build()
	}
	if a.Else != nil {
		children[false] = a.Else.build()
	}
	return selectNode[ConditionResultType]{
		selector: selector[ConditionResultType]{
			check: func(ctx internalCtx) (ConditionResultType, error) {
				val, err := a.If.check(ctx)
				return ConditionResultType(val), err
			},
			string:        a.If.string,
			conditionName: "If",
		},
		children: children,
	}
}

type SimpleAction struct {
	Run   func(Context) error
	Label string
}

func (s SimpleAction) build() node {
	return leafNode{
		action: func(c internalCtx) error {
			return s.Run(c.publicCtx)
		},
		description: s.Label,
	}
}

func Func(label string, fn func(Context) error) Object {
	return SimpleAction{
		Run:   fn,
		Label: label,
	}
}

type scope struct {
	label string
	fn    func() Object
}

func (s scope) build() node {
	return scopeNode{
		nodeProvider: s.fn().build,
		label:        s.label,
	}
}

func Scope(name string, fn func() Object) Object {
	return scope{name, fn}
}

func Nop() Object {
	return SimpleAction{
		Run:   func(Context) error { return nil },
		Label: "nop",
	}
}

func Err(err error) Object {
	return SimpleAction{
		Run:   func(Context) error { return err },
		Label: err.Error(),
	}
}

func newCtx() internalCtx {
	p := printer.New()
	return internalCtx{
		printer: &actionPrinter{
			printFn: p.PersistentPrintln,
			resetFn: p.ResetActionBuffer,
		},
		publicCtx: Context{
			Stdout: p.ActionStdout(),
			Stderr: p.ActionStderr(),
		},
	}
}

func Run(o Object) error {
	ctx := newCtx()
	if err := o.build().run(ctx); err != nil {
		return err
	}
	return nil
}

func RunSilent(o Object) error {
	ctx := newCtx()
	ctx.printer.printFn = func(s string) {}

	return o.build().run(ctx)
}
