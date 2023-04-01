package action

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type not struct {
	Condition
}

func Not(c Condition) Condition {
	return not{Condition: c}
}

func (a not) check(ctx actionCtx) (bool, error) {
	result, resultErr := a.Condition.check(ctx)
	return !result, resultErr
}

func (a not) build() node {
	return node{
		children: []node{a.Condition.build()},
	}
}

func (a not) string() string {
	return fmt.Sprintf("!%s", a.Condition.string())
}

type Const bool

func (a Const) check(ctx actionCtx) (bool, error) {
	return bool(a), nil
}

func (a Const) build() node {
	return node{}
}

func (a Const) string() string {
	return fmt.Sprint(bool(a))
}

type or struct {
	cond1 Condition
	cond2 Condition
}

func Or(c1 Condition, c2 Condition) Condition {
	return or{cond1: c1, cond2: c2}
}

func (a or) build() node {
	return node{
		children: []node{a.cond1.build(), a.cond2.build()},
	}
}

func (a or) check(ctx actionCtx) (bool, error) {
	r1, err1 := a.cond1.check(ctx)
	if err1 != nil {
		return false, err1
	}
	if r1 {
		return true, nil
	}
	r2, err2 := a.cond2.check(ctx)
	if err2 != nil {
		return false, err2
	}
	return r2, nil
}

func (a or) string() string {
	return fmt.Sprintf("%s || %s", a.cond1.string(), a.cond2.string())
}

type and struct {
	cond1 Condition
	cond2 Condition
}

func And(c1 Condition, c2 Condition) Condition {
	return and{cond1: c1, cond2: c2}
}

func (a and) build() node {
	return node{
		children: []node{a.cond1.build(), a.cond2.build()},
	}
}

func (a and) check(ctx actionCtx) (bool, error) {
	r1, err1 := a.cond1.check(ctx)
	if err1 != nil {
		return false, err1
	}
	if !r1 {
		return false, nil
	}
	r2, err2 := a.cond2.check(ctx)
	if err2 != nil {
		return false, err2
	}
	return r1 && r2, nil
}

func (a and) string() string {
	return fmt.Sprintf("%s && %s", a.cond1.string(), a.cond2.string())
}

type SimpleConditionBuilder[T any] struct {
	String          func(T) string
	CreateCondition func(T) func() (bool, error)
}

type SimpleCondition struct {
	checkImpl   func() (bool, error)
	description string
}

func (s SimpleCondition) check(ctx actionCtx) (bool, error) {
	return s.checkImpl()
}

func (s SimpleCondition) build() node {
	return node{}
}

func (a SimpleCondition) string() string {
	return a.description
}

func (s SimpleConditionBuilder[T]) Init() func(T) Condition {
	return func(t T) Condition {
		description := ""
		if s.String != nil {
			description = s.String(t)
		} else {
			description = strings.TrimRight(spew.Sdump(t), "\n ")
		}
		return SimpleCondition{
			checkImpl:   s.CreateCondition(t),
			description: description,
		}
	}
}

var FuncCond = SimpleConditionBuilder[func() (bool, error)]{
	CreateCondition: func(fn func() (bool, error)) func() (bool, error) {
		return func() (bool, error) {
			return fn()
		}
	},
}.Init()
