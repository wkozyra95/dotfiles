package action

type node interface {
	run(actionCtx) error
}

type selector[T any] struct {
	check         func(ctx actionCtx) (T, error)
	string        func() string
	conditionName string
}

type listNode struct {
	children []node
}

func (n listNode) run(ctx actionCtx) error {
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

func (n selectNode[T]) run(ctx actionCtx) error {
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
	action      func(actionCtx) error
	description string
}

func (n leafNode) run(ctx actionCtx) error {
	ctx.printer.printLeafNode(n)
	return n.action(ctx)
}

func (n leafNode) String() string {
	return n.description
}

type wrappedNode struct {
	child         node
	wrapper       func(ctx actionCtx, innerNode node) error
	optionalLabel string
}

func (n wrappedNode) run(ctx actionCtx) error {
	if n.optionalLabel != "" {
		ctx.printer.startScope(n.optionalLabel)
		defer ctx.printer.endScope()
	}
	return n.wrapper(ctx, n.child)
}

type scopeNode struct {
	nodeProvider func(ctx actionCtx) node
	label        string
}

func (n scopeNode) run(ctx actionCtx) error {
	ctx.printer.startScope(n.label)
	defer ctx.printer.endScope()
	build := n.nodeProvider(ctx)
	return build.run(ctx)
}
