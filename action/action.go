package action

import (
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var log = logger.NamedLogger("action")

type actionCtx struct {
	printer *printer
}

type Condition interface {
	check(actionCtx) (bool, error)
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
		wrapper: func(ctx actionCtx, innerNode node) error {
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
	var thenBranch node
	var elseBranch node
	if a.Then != nil {
		thenBranch = a.Then.build()
	}
	if a.Else != nil {
		elseBranch = a.Else.build()
	}
	return selectNode[ConditionResultType]{
		selector: selector[ConditionResultType]{
			check: func(ctx actionCtx) (ConditionResultType, error) {
				val, err := a.If.check(ctx)
				return ConditionResultType(val), err
			},
			string:        a.If.string,
			conditionName: "If",
		},
		children: map[ConditionResultType]node{
			true:  thenBranch,
			false: elseBranch,
		},
	}
}

type SimpleAction struct {
	Run   func() error
	Label string
}

func (s SimpleAction) build() node {
	return leafNode{
		action: func(ac actionCtx) error {
			return s.Run()
		},
		description: s.Label,
	}
}

func Func(label string, fn func() error) Object {
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
		nodeProvider: func(ctx actionCtx) node {
			return s.fn().build()
		},
		label: s.label,
	}
}

func Scope(name string, fn func() Object) Object {
	return scope{name, fn}
}

func Nop() Object {
	return SimpleAction{
		Run:   func() error { return nil },
		Label: "nop",
	}
}

func Err(err error) Object {
	return SimpleAction{
		Run:   func() error { return err },
		Label: err.Error(),
	}
}

func Run(o Object) error {
	return o.build().run(actionCtx{printer: &printer{
		printFn: func(s string) { println(s) },
	}})
}

func RunSilent(o Object) error {
	return o.build().run(actionCtx{printer: &printer{
		printFn: func(s string) {},
	}})
}
