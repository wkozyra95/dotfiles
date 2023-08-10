package action

type node interface {
	run(internalCtx) error
}

type selector[T any] struct {
	check         func(ctx internalCtx) (T, error)
	string        func() string
	conditionName string
}

type listNode struct {
	children []node
}

func (n listNode) run(ctx internalCtx) error {
	ctx.printer.startList()
	defer ctx.printer.endList()
	for _, node := range n.children {
		if err := node.run(ctx); err != nil {
			return err
		}
	}
	return nil
}

type conditionalKey interface {
	comparable
	String() string
}

type selectNode[T conditionalKey] struct {
	selector selector[T]
	children map[T]node
}

func (n selectNode[T]) run(ctx internalCtx) error {
	startSelectNodeCondition(ctx.printer, n)
	selected, selectErr := n.selector.check(ctx)
	if selectErr != nil {
		return selectErr
	}
	selectedAction := n.children[selected]
	if selectedAction != nil {
		ctx.printer.startSelectNodeBranch(selected.String())
		defer ctx.printer.endSelectNodeBranch()
		return selectedAction.run(ctx)
	} else {
		ctx.printer.printSelectNodeNoBranchSelected(selected.String())
	}
	return nil
}

type leafNode struct {
	action      func(internalCtx) error
	description string
}

func (n leafNode) run(ctx internalCtx) error {
	ctx.printer.printLeafNode(n)
	return n.action(ctx)
}

func (n leafNode) String() string {
	return n.description
}

type wrappedNode struct {
	child         node
	wrapper       func(ctx internalCtx, innerNode node) error
	optionalLabel string
}

func (n wrappedNode) run(ctx internalCtx) error {
	if n.optionalLabel != "" {
		ctx.printer.startScope(n.optionalLabel)
		defer ctx.printer.endScope()
	}
	return n.wrapper(ctx, n.child)
}

type scopeNode struct {
	nodeProvider func() node
	label        string
}

func (n scopeNode) run(ctx internalCtx) error {
	ctx.printer.startScope(n.label)
	defer ctx.printer.endScope()
	build := n.nodeProvider()
	return build.run(ctx)
}
