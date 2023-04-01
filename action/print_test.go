package action

import (
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
 - If: true
   Then: example one line output
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
   Then: example one line output
 - first line
   second line
   third line`

var fakeAction = SimpleActionBuilder[string]{
	CreateRun: func(args string) func() error {
		return func() error {
			return nil
		}
	},
	String: func(args string) string {
		return args
	},
}.Init()

func TestNodePrint(t *testing.T) {
	err := List{
		fakeAction(oneliner),
		fakeAction(multiliner),
		fakeAction(oneliner),
		WithCondition{
			If:   Const(true),
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
	}.run(actionCtx{}, 0)
	assert.Equal(t, nil, err)
}
