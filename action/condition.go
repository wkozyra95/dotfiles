package action

import (
	"fmt"
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

func (a not) string() string {
	return fmt.Sprintf("!%s", a.Condition.string())
}

type Const bool

func (a Const) check(ctx actionCtx) (bool, error) {
	return bool(a), nil
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

type SimpleCondition struct {
	Check func() (bool, error)
	Label string
}

func (s SimpleCondition) check(ctx actionCtx) (bool, error) {
	return s.Check()
}

func (a SimpleCondition) string() string {
	return a.Label
}

func FuncCond(label string, fn func() (bool, error)) Condition {
	return SimpleCondition{
		Check: func() (bool, error) {
			return fn()
		},
		Label: label,
	}
}

func LabeledConst(labelName string, value bool) Condition {
	label := labelName
	if !value {
		label = fmt.Sprintf("!%s", labelName)
	}
	return SimpleCondition{
		Check: func() (bool, error) {
			return value, nil
		},
		Label: label,
	}
}
