package action

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	oneliner   = "example one line output"
	multiliner = `first line
second line
third line`
)

var expectedOutput = ` - example one line output
 - first line
   second line
   third line
 - example one line output
 - If: !!true
   Then:
     example one line output
 - If: true
   Then:
     first line
     second line
     third line
 - If: true
   Then:
     - first line
       second line
       third line
     - first line
       second line
       third line
 - If: true
   Then:
     example one line output
 - first line
   second line
   third line
 - With single action:
     first line
     second line
     third line
 - With list:
     - example one line output
     - first line
       second line
       third line
 - With condition:
     If: true
     Then:
       - example one line output
       - example one line output`

func fakeAction(text string) Object {
	return SimpleAction{
		Run:   func() error { return nil },
		Label: text,
	}
}

func TestNodePrint(t *testing.T) {
	actions := List{
		fakeAction(oneliner),
		fakeAction(multiliner),
		fakeAction(oneliner),
		WithCondition{
			If:   Not(Not(Const(true))),
			Then: fakeAction(oneliner),
		},
		WithCondition{
			If:   Const(true),
			Then: fakeAction(multiliner),
		},
		WithCondition{
			If: Const(true),
			Then: List{
				fakeAction(multiliner),
				fakeAction(multiliner),
			},
		},
		List{
			WithCondition{
				If:   Const(true),
				Then: fakeAction(oneliner),
			},
			List{
				List{
					fakeAction(multiliner),
				},
			},
		},
		Scope("With single action", func() Object {
			return fakeAction(multiliner)
		}),
		Scope("With list", func() Object {
			return List{
				fakeAction(oneliner),
				fakeAction(multiliner),
			}
		}),
		Scope("With condition", func() Object {
			return WithCondition{
				If: Const(true),
				Then: List{
					fakeAction(oneliner),
					fakeAction(oneliner),
				},
			}
		}),
	}

	result := []string{}
	ctx, err := newCtx(false)
	assert.Nil(t, err)
	ctx.printer.printFn = func(s string) {
		result = append(result, s)
	}
	ctx.printer.stack = []string{" "}

	runErr := actions.build().run(ctx)
	assert.Nil(t, runErr)
	assert.Equal(t, expectedOutput, strings.Join(result, "\n"))
}
