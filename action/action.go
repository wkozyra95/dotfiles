package action

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var log = logger.NamedLogger("action")

func printAction(depth int, text string) {
	split := strings.Split(text, "\n")
	for i := range split {
		split[i] = fmt.Sprintf("%s%s", strings.Repeat("   ", depth), split[i])
	}
	fmt.Println(strings.Join(split, "\n"))
}

type actionCtx struct {
	print bool
}

type Condition interface {
	check(ctx actionCtx) (bool, error)
	build() node
	string() string
}

type Object interface {
	run(ctx actionCtx, depth int) error
	build() node
	string() string
}

type node struct {
	children []node
}

type List []Object

func (l List) build() node {
	children := []node{}
	for _, child := range l {
		children = append(children, child.build())
	}
	return node{
		children: children,
	}
}

func (l List) run(ctx actionCtx, depth int) error {
	for _, action := range l {
		if ctx.print {
			lines := strings.Split(action.string(), "\n")
			if action.string() == "condition" {
				// empty
			} else if len(lines) == 1 {
				printAction(depth, fmt.Sprintf(" - %s", lines[0]))
			} else {
				printAction(depth, fmt.Sprintf(" - %s", lines[0]))
				printAction(depth+1, strings.Join(lines[1:], "\n"))
			}
		}
		err := action.run(ctx, depth+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l List) string() string {
	return ""
}

type Optional struct {
	Object Object
}

func (o Optional) build() node {
	return node{
		children: []node{o.Object.build()},
	}
}

func (o Optional) run(ctx actionCtx, depth int) error {
	err := o.Object.run(ctx, depth+1)
	if err != nil {
		log.Error(err)
		if !prompt.ConfirmPrompt("Install failed, do you want to continue?") {
			return err
		}
	}
	return nil
}

func (o Optional) string() string {
	return ""
}

type WithCondition struct {
	If   Condition
	Then Object
	Else Object
}

func (a WithCondition) run(ctx actionCtx, depth int) error {
	if ctx.print {
		// This is hack, build action tree and print based on that
		printAction(depth-1, fmt.Sprintf(" - If: %s", a.If.string()))
	}
	result, err := a.If.check(ctx)
	if err != nil {
		return err
	}
	if result {
		if ctx.print {
			printAction(depth, fmt.Sprintf("Then: %s", a.Then.string()))
		}
		return a.Then.run(ctx, depth+1)
	} else if a.Else != nil {
		if ctx.print {
			printAction(depth, fmt.Sprintf("Else: %s", a.Else.string()))
		}
		return a.Else.run(ctx, depth+1)
	} else {
		if ctx.print {
			printAction(depth, fmt.Sprintf("Else: do nothing"))
		}
	}
	return nil
}

func (a WithCondition) build() node {
	return node{
		children: []node{
			a.If.build(),
			a.Then.build(),
			a.Else.build(),
		},
	}
}

func (a WithCondition) string() string {
	return "condition"
}

type SimpleActionBuilder[T any] struct {
	CreateRun func(T) func() error
	String    func(T) string
}

type SimpleAction struct {
	runImpl     func() error
	description string
}

func (s SimpleAction) run(ctx actionCtx, depth int) error {
	return s.runImpl()
}

func (s SimpleAction) build() node {
	return node{}
}

func (a SimpleAction) string() string {
	return a.description
}

func (s SimpleActionBuilder[T]) Init() func(T) Object {
	return func(t T) Object {
		description := ""
		if s.String != nil {
			description = s.String(t)
		} else {
			description = strings.TrimRight(spew.Sdump(t), "\n ")
		}
		return SimpleAction{
			runImpl:     s.CreateRun(t),
			description: description,
		}
	}
}

var Func = SimpleActionBuilder[func() error]{
	CreateRun: func(fn func() error) func() error {
		return func() error {
			return fn()
		}
	},
}.Init()

type scope struct {
	fn func() Object
}

func (s scope) run(ctx actionCtx, depth int) error {
	return s.fn().run(ctx, depth)
}

func (s scope) build() node {
	return s.fn().build()
}

func (a scope) string() string {
	return ""
}

func Scope(fn func() Object) Object {
	return scope{fn}
}

var nop = SimpleActionBuilder[struct{}]{
	CreateRun: func(ignored struct{}) func() error {
		return func() error {
			return nil
		}
	},
}.Init()

func Nop() Object {
	return nop(struct{}{})
}

var errAction = SimpleActionBuilder[error]{
	CreateRun: func(err error) func() error {
		return func() error {
			return err
		}
	},
}.Init()

func Err(err error) Object {
	return errAction(err)
}

func Run(o Object) error {
	return o.run(actionCtx{print: true}, 0)
}

func RunSilent(o Object) error {
	return o.run(actionCtx{print: false}, 0)
}
